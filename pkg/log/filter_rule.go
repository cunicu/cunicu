// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-FileCopyrightText: 2020 Manfred Touron <https://manfred.life>
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"fmt"
	"math"
	"path"
	"strings"
	"sync"
	"sync/atomic"

	"go.uber.org/zap/zapcore"
)

// FilterFunc is used to check whether to filter the given entry and filters out.
type FilterFunc func(zapcore.Entry) bool

type FilterRule struct {
	Filter     FilterFunc
	Expression string
	Level      Level
}

type AtomicFilterRule struct {
	atomic.Pointer[FilterRule]
}

func NewFilterRule(ff FilterFunc) *FilterRule {
	return &FilterRule{
		Filter: ff,
		Level:  Level(ff.Level()),
	}
}

func ParseFilterRule(expr string) (*FilterRule, error) {
	ff, err := ParseRules(expr)
	if err != nil {
		return nil, err
	}

	return &FilterRule{
		Filter:     ff,
		Level:      Level(ff.Level()),
		Expression: expr,
	}, nil
}

func (f FilterFunc) Level() zapcore.Level {
	for l := MinLevel; l <= MaxLevel; l++ {
		if f(zapcore.Entry{
			Level: zapcore.Level(l),
		}) {
			return zapcore.Level(l)
		}
	}

	return zapcore.InvalidLevel
}

// ByNamespaces takes a list of patterns to filter out logs based on their namespaces.
// Patterns are checked using path.Match.
func ByNamespaces(input string) FilterFunc { //nolint:gocognit
	if input == "" {
		return AlwaysFalseFilter
	}
	patterns := strings.Split(input, ",")

	// Edge case optimization (always true)
	{
		hasIncludeWildcard := false
		hasExclude := false
		for _, pattern := range patterns {
			if pattern == "" {
				continue
			}
			if pattern == "*" {
				hasIncludeWildcard = true
			}
			if pattern[0] == '-' {
				hasExclude = true
			}
		}
		if hasIncludeWildcard && !hasExclude {
			return AlwaysTrueFilter
		}
	}

	var mutex sync.Mutex
	matchMap := map[string]bool{}
	return func(entry zapcore.Entry) bool {
		mutex.Lock()
		defer mutex.Unlock()

		if _, found := matchMap[entry.LoggerName]; !found {
			matchMap[entry.LoggerName] = false
			matchInclude := false
			matchExclude := false
			for _, pattern := range patterns {
				switch {
				case pattern[0] == '-' && !matchExclude:
					if matched, _ := path.Match(pattern[1:], entry.LoggerName); matched {
						matchExclude = true
					}
				case pattern[0] != '-' && !matchInclude:
					if matched, _ := path.Match(pattern, entry.LoggerName); matched {
						matchInclude = true
					}
				}
			}
			matchMap[entry.LoggerName] = matchInclude && !matchExclude
		}
		return matchMap[entry.LoggerName]
	}
}

// ExactLevel filters out entries with an invalid level.
func ExactLevel(level Level) FilterFunc {
	return func(entry zapcore.Entry) bool {
		return Level(entry.Level) == level
	}
}

// MinimumLevel filters out entries with a too low level.
func MinimumLevel(level Level) FilterFunc {
	return func(entry zapcore.Entry) bool {
		return Level(entry.Level) >= level
	}
}

// Any checks if any filter returns true.
func Any(filters ...FilterFunc) FilterFunc {
	return func(entry zapcore.Entry) bool {
		for _, filter := range filters {
			if filter == nil {
				continue
			}
			if filter(entry) {
				return true
			}
		}
		return false
	}
}

// Reverse checks is the passed filter returns false.
func Reverse(filter FilterFunc) FilterFunc {
	return func(entry zapcore.Entry) bool {
		return !filter(entry)
	}
}

// All checks if all filters return true.
func All(filters ...FilterFunc) FilterFunc {
	return func(entry zapcore.Entry) bool {
		var atLeastOneSuccessful bool
		for _, filter := range filters {
			if filter == nil {
				continue
			}
			if !filter(entry) {
				return false
			}
			atLeastOneSuccessful = true
		}
		return atLeastOneSuccessful
	}
}

// ParseRules takes a CLI-friendly set of rules to construct a filter.
//
// Syntax
//
//	pattern: RULE [RULE...]
//	RULE: one of:
//	 - LEVELS:NAMESPACES
//	 - LEVELS
//	LEVELS: LEVEL,[,LEVEL]
//	LEVEL: see `Level Patterns`
//	NAMESPACES: NAMESPACE[,NAMESPACE]
//	NAMESPACE: one of:
//	 - namespace     // Should be exactly this namespace
//	 - *mat*ch*      // Should match
//	 - -NAMESPACE    // Should not match
//
// Examples
//
//	*:*                          everything
//	info:*                       level info;  any namespace
//	info+:*                      levels info, warn, error, dpanic, panic, and fatal; any namespace
//	info,warn:*                  levels info, warn; any namespace
//	ns1                          any level; namespace 'ns1'
//	*:ns1                        any level; namespace 'ns1'
//	ns1*                         any level; namespaces matching 'ns1*'
//	*:ns1*                       any level; namespaces matching 'ns1*'
//	*:ns1,ns2                    any level; namespaces 'ns1' and 'ns2'
//	*:ns*,-ns3*                  any level; namespaces matching 'ns*' but not matching 'ns3*'
//	info:ns1                     level info; namespace 'ns1'
//	info,warn:ns1,ns2            levels info and warn; namespaces 'ns1' and 'ns2'
//	info:ns1 warn:n2             level info + namespace 'ns1' OR level warn and namespace 'ns2'
//	info,warn:myns* error+:*     levels info or warn and namespaces matching 'myns*' OR levels error, dpanic, panic or fatal for any namespace
func ParseRules(expr string) (FilterFunc, error) {
	var filters []FilterFunc

	// Rules are separated by spaces, tabs or \n
	for _, rule := range strings.Fields(expr) {
		// Split rule into parts (separated by ':')
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}
		parts := strings.SplitN(rule, ":", 2)
		var left, right string
		switch len(parts) {
		case 1:
			left = parts[0] // If no separator, right matches everything
			right = "*"
		case 2:
			if parts[0] == "" || parts[1] == "" {
				return nil, ErrBadSyntax
			}
			left = parts[0]
			right = parts[1]
		default:
			return nil, ErrBadSyntax
		}

		levelFilter, err := ByLevels(left)
		if err != nil {
			return nil, err
		}
		namespaceFilter := ByNamespaces(right)

		filters = append(filters, All(levelFilter, namespaceFilter))
	}

	return Any(filters...), nil
}

// ByLevels creates a FilterFunc based on a pattern.
//
// Level Patterns
//
//	| Pattern | Debug(X) | Debug | Info | Warn | Error | DPanic | Panic | Fatal |
//	| ------- | -----    | ----- | ---- | ---- | ----- | ------ | ----- | ----- |
//	| <empty> | X        | X     | X    | X    | X     | X      | X     | X     |
//	| *       | X        | X     | X    | X    | X     | x      | X     | X     |
//	| =debugX | X        |       |      |      |       |        |       |       |
//	| =debug  |          | X     |      |      |       |        |       |       |
//	| =info   |          |       | X    |      |       |        |       |       |
//	| =warn   |          |       |      | X    |       |        |       |       |
//	| =error  |          |       |      |      | X     |        |       |       |
//	| =dpanic |          |       |      |      |       | X      |       |       |
//	| =panic  |          |       |      |      |       |        | X     |       |
//	| =fatal  |          |       |      |      |       |        |       | X     |
//	| debug   | X        | X     | X    | x    | X     | X      | X     | X     |
//	| info    |          |       | X    | X    | X     | X      | X     | X     |
//	| warn    |          |       |      | X    | X     | X      | X     | X     |
//	| error   |          |       |      |      | X     | X      | X     | X     |
//	| dpanic  |          |       |      |      |       | X      | X     | X     |
//	| panic   |          |       |      |      |       |        | X     | X     |
//	| fatal   |          |       |      |      |       |        |       | X     |
func ByLevels(pattern string) (FilterFunc, error) {
	var enabled uint64
	for _, part := range strings.Split(pattern, ",") {
		if part == "" || part == "*" {
			enabled |= math.MaxUint64
		} else {
			op := part[0]
			switch op {
			case '=', '<', '>':
				part = part[1:]
			default:
				op = '>'
			}

			var level Level
			if err := level.UnmarshalText([]byte(part)); err != nil {
				return nil, fmt.Errorf("%w: %s", ErrUnsupportedKeyword, part)
			}

			bit := levelToBit(level)

			switch op {
			case '=':
				enabled |= bit
			case '<':
				enabled |= bit
				enabled |= ^(bit - 1)
			case '>':
				enabled |= bit
				enabled |= bit - 1
			}
		}
	}

	return func(e zapcore.Entry) bool {
		return levelToBit(Level(e.Level))&enabled != 0
	}, nil
}

func levelToBit(l Level) uint64 {
	return 1 << (Level(zapcore.FatalLevel) - l)
}

// MustParseRules calls ParseRules and panics if initialization failed.
func MustParseRules(expr string) FilterFunc {
	filter, err := ParseRules(expr)
	if err != nil {
		panic(err)
	}
	return filter
}

func AlwaysFalseFilter(_ zapcore.Entry) bool {
	return false
}

func AlwaysTrueFilter(_ zapcore.Entry) bool {
	return true
}
