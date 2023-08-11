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

var filter atomic.Pointer[Filter] //nolint:gochecknoglobals

// FilterFunc is used to check whether to filter the given entry and filters out.
type FilterFunc func(zapcore.Entry) bool

type Filter struct {
	FilterFunc

	Rules []string // The list of rules used to construct the filter
	Level Level    // The highest level which the filter lets pass-through
}

func (f *Filter) String() string {
	return strings.Join(f.Rules, ",")
}

func UpdateFilter(f *Filter) {
	filter.Store(f)
}

func CurrentFilter() *Filter {
	return filter.Load()
}

func ParseFilter(rules []string) (*Filter, error) {
	ff, err := ParseRules(rules)
	if err != nil {
		return nil, err
	}

	return &Filter{
		FilterFunc: ff,
		Level:      Level(ff.Level()),
		Rules:      rules,
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

// Any checks if any filter returns true.
func Any(filterFuncs ...FilterFunc) FilterFunc {
	return func(entry zapcore.Entry) bool {
		for _, filter := range filterFuncs {
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
func Reverse(filterFunc FilterFunc) FilterFunc {
	return func(entry zapcore.Entry) bool {
		return !filterFunc(entry)
	}
}

// All checks if all filters return true.
func All(filterFuncs ...FilterFunc) FilterFunc {
	return func(entry zapcore.Entry) bool {
		var atLeastOneSuccessful bool
		for _, filter := range filterFuncs {
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
//	rules: slice of RULE
func ParseRules(rules []string) (FilterFunc, error) {
	var filterFuncs []FilterFunc

	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}

		filter, err := ParseRule(rule)
		if err != nil {
			return nil, err
		}

		filterFuncs = append(filterFuncs, filter)
	}

	return Any(filterFuncs...), nil
}

// ParseRules takes a CLI-friendly set of rules to construct a filter.
// Syntax
//
//	rule: RULE
//	RULE: one of:
//	 - LEVELS:NAMESPACES
//	 - LEVELS
//	LEVELS: LEVEL[,LEVELS]
//	LEVEL: see `Level Patterns`
//	NAMESPACES: NAMESPACE[,NAMESPACES]
//	NAMESPACE: one of:
//	 - namespace     // Should be exactly this namespace
//	 - *mat*ch*      // Should match
//	 - -NAMESPACE    // Should not match
func ParseRule(rule string) (FilterFunc, error) {
	// Split rule into parts (separated by ':')
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

	levelFilterFunc, err := ByLevels(left)
	if err != nil {
		return nil, err
	}

	namespaceFilterFunc := ByNamespaces(right)

	return All(levelFilterFunc, namespaceFilterFunc), nil
}

// ByLevels creates a FilterFunc based on a pattern.
//
// Syntax
//
//	pattern: LEVELS
//	LEVELS: LEVEL[,LEVELS]
//	LEVEL: one of
//	 -  SEVERITY for matching all levels with equal or higher severity
//	 - >SEVERITY for matching all levels with equal or higher severity
//	 - =SEVERITY for matching all levels with equal severity
//	 - <SEVERITY for matching all levels with lower severity
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

func AlwaysFalseFilter(_ zapcore.Entry) bool {
	return false
}

func AlwaysTrueFilter(_ zapcore.Entry) bool {
	return true
}
