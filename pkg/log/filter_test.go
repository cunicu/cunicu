// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-FileCopyrightText: 2020 Manfred Touron <https://manfred.life>
// SPDX-License-Identifier: Apache-2.0

package log_test

import (
	"errors"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/onsi/gomega/gcustom"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"cunicu.li/cunicu/pkg/log"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	errMismatchingLogMessage = errors.New("mismatch in log message")
	errMismatchingLoggerName = errors.New("mismatch in logger name")
	errMismatchingLogLevel   = errors.New("mismatch in log level")
	errMismatchingField      = errors.New("mismatch in context field")
)

// mustParseRule calls ParseRules and panics if initialization failed.
func mustParseRules(rules ...string) log.FilterFunc {
	filter, err := log.ParseRules(rules)
	if err != nil {
		panic(err)
	}

	return filter
}

func matchEntry(expectedEntry zapcore.Entry, expectedFields ...zapcore.Field) OmegaMatcher {
	return gcustom.MakeMatcher(func(actualEntry observer.LoggedEntry) (bool, error) {
		if expectedEntry.Message != "" && expectedEntry.Message != actualEntry.Message {
			return false, fmt.Errorf("%w: %s != %s", errMismatchingLogMessage, expectedEntry.Message, actualEntry.Message)
		}

		if expectedEntry.LoggerName != "" && expectedEntry.LoggerName != actualEntry.LoggerName {
			return false, fmt.Errorf("%w: %s != %s", errMismatchingLoggerName, expectedEntry.LoggerName, actualEntry.LoggerName)
		}

		if expectedEntry.Level != zapcore.InfoLevel && expectedEntry.Level != actualEntry.Level {
			return false, fmt.Errorf("%w: %s != %s", errMismatchingLogLevel, expectedEntry.Level, actualEntry.Level)
		}

		if len(expectedFields) == 0 {
			return true, nil
		}

		expectedFieldMap := map[string]zapcore.Field{}
		for _, field := range expectedFields {
			expectedFieldMap[field.Key] = field
		}

		for _, actualField := range actualEntry.Context {
			if expectedField, ok := expectedFieldMap[actualField.Key]; ok && !expectedField.Equals(actualField) {
				return false, fmt.Errorf("%w: %s", errMismatchingField, actualField.Key)
			}
		}

		return true, nil
	})
}

func makeLogger(filterFunc log.FilterFunc) (*zap.Logger, *observer.ObservedLogs) {
	observed, logs := observer.New(zapcore.DebugLevel)

	filter := new(atomic.Pointer[log.Filter])
	filter.Store(&log.Filter{
		FilterFunc: filterFunc,
		Level:      log.Level(filterFunc.Level()),
	})

	filtered := log.NewFilteredCore(observed, filter)

	return zap.New(filtered), logs
}

// CheckAnyLevel determines whether at least one log level isn't filtered-out by the logger.
func checkAnyLevel(logger *zap.Logger) bool {
	for l := log.MinLevel; l <= log.DPanicLevel; l++ {
		if checkLevel(logger, l) {
			return true
		}
	}

	return false
}

// CheckLevel determines whether a specific log level would produce log or not.
func checkLevel(logger *zap.Logger, level log.Level) bool {
	return logger.Check(zapcore.Level(level), "") != nil
}

var _ = Describe("filter", func() {
	Describe("NewFilteredCore", func() {
		It("wrap", func() {
			next, logs := observer.New(zapcore.DebugLevel)
			logger := zap.New(next)
			defer logger.Sync() //nolint:errcheck

			filterFunc := mustParseRules("*:demo*")
			filter := new(atomic.Pointer[log.Filter])
			filter.Store(&log.Filter{
				FilterFunc: filterFunc,
				Level:      log.Level(filterFunc.Level()),
			})

			logger = logger.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
				return log.NewFilteredCore(c, filter)
			}))

			logger.Debug("hello world!")
			logger.Named("demo").Debug("hello earth!")
			logger.Named("other").Debug("hello universe!")

			Expect(logs.All()).To(HaveExactElements(
				matchEntry(zapcore.Entry{Message: "hello earth!"}),
			))
		})

		It("new logger", func() {
			logger, logs := makeLogger(mustParseRules("*:demo*"))
			defer logger.Sync() //nolint:errcheck

			logger.Debug("hello world!")
			logger.Named("demo").Debug("hello earth!")
			logger.Named("other").Debug("hello universe!")

			Expect(logs.All()).To(HaveExactElements(
				matchEntry(zapcore.Entry{Message: "hello earth!"}),
			))
		})
	})

	Describe("FilterFunc", func() {
		It("ByNamespace", func() {
			logger, logs := makeLogger(log.ByNamespaces("demo1.*,demo3.*"))
			defer logger.Sync() //nolint:errcheck

			logger.Debug("hello city!")
			logger.Named("demo1.frontend").Debug("hello region!")
			logger.Named("demo2.frontend").Debug("hello planet!")
			logger.Named("demo3.frontend").Debug("hello solar system!")

			Expect(logs.All()).To(HaveExactElements(
				matchEntry(zapcore.Entry{Message: "hello region!", LoggerName: "demo1.frontend", Level: zapcore.DebugLevel}),
				matchEntry(zapcore.Entry{Message: "hello solar system!", LoggerName: "demo3.frontend", Level: zapcore.DebugLevel}),
			))
		})

		It("custom", func() {
			logger, logs := makeLogger(func(entry zapcore.Entry) bool {
				return strings.Contains(entry.Message, "planet")
			})
			defer logger.Sync() //nolint:errcheck

			logger.Debug("hello city!")
			logger.Debug("hello region!")
			logger.Debug("hello planet!")
			logger.Debug("hello solar system!")
			logger.Debug("hello universe!")
			logger.Debug("hello multiverse!")

			Expect(logs.All()).To(HaveExactElements(
				matchEntry(zapcore.Entry{Message: "hello planet!"}),
			))
		})

		DescribeTable("simple",
			func(
				filterFunc log.FilterFunc,
				expectedLogs []string,
			) {
				logger, logs := makeLogger(filterFunc)
				defer logger.Sync() //nolint:errcheck

				logger.Debug("a")
				logger.Info("b")
				logger.Warn("c")
				logger.Error("d")

				gotLogs := []string{}
				for _, log := range logs.All() {
					gotLogs = append(gotLogs, log.Message)
				}

				Expect(gotLogs).To(Equal(expectedLogs))
			},
			Entry("allow-all",
				func(_ zapcore.Entry) bool {
					return true
				},
				[]string{"a", "b", "c", "d"},
			),
			Entry("disallow-all",
				func(_ zapcore.Entry) bool {
					return false
				},
				[]string{},
			),
		)
	})

	const (
		allDebug   = "aeimquy2"
		allInfo    = "bfjnrvz3"
		allWarn    = "cgkosw04"
		allError   = "dhlptx15"
		everything = "abcdefghijklmnopqrstuvwxyz012345"
	)

	It("ParseRule", func() {
		// *:myns             => any level, myns namespace
		// info,warn:myns.*   => info or warn level, any namespace matching myns.*
		// error:*            => everything with error level
		logger, logs := makeLogger(mustParseRules("*:myns", "info,warn:myns.*", "error:*"))
		defer logger.Sync() //nolint:errcheck

		logger.Debug("top debug") // No match
		Expect(logs.TakeAll()).To(BeEmpty())

		logger.Named("myns").Debug("myns debug") // Matches *:myns
		Expect(logs.TakeAll()).To(HaveExactElements(
			matchEntry(zapcore.Entry{Message: "myns debug", LoggerName: "myns", Level: zapcore.DebugLevel}),
		))

		logger.Named("bar").Debug("bar debug") // No match
		Expect(logs.TakeAll()).To(BeEmpty())

		logger.Named("myns").Named("foo").Debug("myns.foo debug") // No match
		Expect(logs.TakeAll()).To(BeEmpty())

		logger.Info("top info") // No match
		Expect(logs.TakeAll()).To(BeEmpty())

		logger.Named("myns").Info("myns info") // Matches *:myns
		Expect(logs.TakeAll()).To(HaveExactElements(
			matchEntry(zapcore.Entry{Message: "myns info", LoggerName: "myns", Level: zapcore.InfoLevel}),
		))

		logger.Named("bar").Info("bar info") // No match
		Expect(logs.TakeAll()).To(BeEmpty())

		logger.Named("myns").Named("foo").Info("myns.foo info") // Matches info,warn:myns.*
		Expect(logs.TakeAll()).To(HaveExactElements(
			matchEntry(zapcore.Entry{Message: "myns.foo info", LoggerName: "myns.foo", Level: zapcore.InfoLevel}),
		))

		logger.Warn("top warn") // No match
		Expect(logs.TakeAll()).To(BeEmpty())

		logger.Named("myns").Warn("myns warn") // Matches *:myns
		Expect(logs.TakeAll()).To(HaveExactElements(
			matchEntry(zapcore.Entry{Message: "myns warn", LoggerName: "myns", Level: zapcore.WarnLevel}),
		))

		logger.Named("bar").Warn("bar warn") // No match
		Expect(logs.TakeAll()).To(BeEmpty())

		logger.Named("myns").Named("foo").Warn("myns.foo warn") // Matches info,warn:myns.*
		Expect(logs.TakeAll()).To(HaveExactElements(
			matchEntry(zapcore.Entry{Message: "myns.foo warn", LoggerName: "myns.foo", Level: zapcore.WarnLevel}),
		))

		logger.Error("top error") // Matches error:*
		Expect(logs.TakeAll()).To(HaveExactElements(
			matchEntry(zapcore.Entry{Message: "top error", Level: zapcore.ErrorLevel}),
		))

		logger.Named("myns").Error("myns error") // Matches *:myns and error:*
		Expect(logs.TakeAll()).To(HaveExactElements(
			matchEntry(zapcore.Entry{Message: "myns error", LoggerName: "myns", Level: zapcore.ErrorLevel}),
		))

		logger.Named("bar").Error("bar error") // Matches error:*
		Expect(logs.TakeAll()).To(HaveExactElements(
			matchEntry(zapcore.Entry{Message: "bar error", LoggerName: "bar", Level: zapcore.ErrorLevel}),
		))

		logger.Named("myns").Named("foo").Error("myns.foo error") // Matches error:*
		Expect(logs.TakeAll()).To(HaveExactElements(
			matchEntry(zapcore.Entry{Message: "myns.foo error", LoggerName: "myns.foo", Level: zapcore.ErrorLevel}),
		))
	})

	DescribeTable("ParseRules",
		func(
			expectedLogs string,
			expectedError error,
			rules ...string,
		) {
			filterFunc, err := log.ParseRules(rules)
			if err != nil {
				Expect(err).To(MatchError(expectedError))

				return
			}

			logger, logs := makeLogger(filterFunc)
			defer logger.Sync() //nolint:errcheck

			logger.Debug("a")
			logger.Info("b")
			logger.Warn("c")
			logger.Error("d")

			logger.Named("foo").Debug("e")
			logger.Named("foo").Info("f")
			logger.Named("foo").Warn("g")
			logger.Named("foo").Error("h")

			logger.Named("bar").Debug("i")
			logger.Named("bar").Info("j")
			logger.Named("bar").Warn("k")
			logger.Named("bar").Error("l")

			logger.Named("baz").Debug("m")
			logger.Named("baz").Info("n")
			logger.Named("baz").Warn("o")
			logger.Named("baz").Error("p")

			logger.Named("foo").Named("bar").Debug("q")
			logger.Named("foo").Named("bar").Info("r")
			logger.Named("foo").Named("bar").Warn("s")
			logger.Named("foo").Named("bar").Error("t")

			logger.Named("foo").Named("foo").Debug("u")
			logger.Named("foo").Named("foo").Info("v")
			logger.Named("foo").Named("foo").Warn("w")
			logger.Named("foo").Named("foo").Error("x")

			logger.Named("bar").Named("foo").Debug("y")
			logger.Named("bar").Named("foo").Info("z")
			logger.Named("bar").Named("foo").Warn("0")
			logger.Named("bar").Named("foo").Error("1")

			logger.Named("qux").Named("foo").Debug("2")
			logger.Named("qux").Named("foo").Info("3")
			logger.Named("qux").Named("foo").Warn("4")
			logger.Named("qux").Named("foo").Error("5")

			gotLogs := []string{}
			for _, log := range logs.All() {
				gotLogs = append(gotLogs, log.Message)
			}

			Expect(strings.Join(gotLogs, "")).To(Equal(expectedLogs))
		},
		Entry("empty", "", nil, ""),
		Entry("everything", everything, nil, "*"),
		Entry("debug", everything, nil, "debug:*"),
		Entry("all-debug", allDebug, nil, "=debug:*"),
		Entry("all-info", allInfo, nil, "=info:*"),
		Entry("all-warn", allWarn, nil, "=warn:*"),
		Entry("all-error", allError, nil, "=error:*"),
		Entry("all-info-and-warn-1", "bcfgjknorsvwz034", nil, "=info,=warn:*"),
		Entry("all-info-and-warn-2", "bcfgjknorsvwz034", nil, "=info:*", "=warn:*"),
		Entry("warn", "cdghklopstwx0145", nil, "warn:*"),
		Entry("redundant-1", allInfo, nil, "=info,=info:*", "=info:*"),
		Entry("redundant-2", everything, nil, "*", "*:*", "=info:*"),
		Entry("foo-ns", "efgh", nil, "*:foo"),
		Entry("foo-ns-debug,info", "ef", nil, "=debug,=info:foo"),
		Entry("foo.star-ns", "qrstuvwx", nil, "*:foo.*"),
		Entry("foo.star-ns-debug,info", "qruv", nil, "=debug,=info:foo.*"),
		Entry("all-in-one", "defghjklpqtux15", nil, "*:foo", "=debug:foo.*", "=info,=warn:bar", "error:*"),
		Entry("exclude-1", "fr", nil, "=info:test,foo*,-foo.foo"),
		Entry("exclude-2", "fr", nil, "=info:test,foo*,-*.foo"),
		Entry("exclude-3", "yz012345", nil, "*:test,*.foo,-foo.*"),
		Entry("exclude-4", "abcdmnopqrstuvwxyz012345", nil, "*:*,-foo,-bar"),
		Entry("exclude-5", "efghijklqrst", nil, "*:foo*,bar*,-foo.foo,-bar.foo"),
		Entry("exclude-6", "efghijklqrst", nil, "*:foo*,-foo.foo,bar*,-bar.foo"),
		Entry("invalid-left", "", log.ErrUnsupportedKeyword, "invalid:*"),
		Entry("missing-left", "", log.ErrBadSyntax, ":*"),
		Entry("missing-right", "", log.ErrBadSyntax, ":*"),
		PEntry("missing-exclude-pattern", "", log.ErrBadSyntax, "*:-"),
	)

	Describe("Check", func() {
		DescribeTable("simple",
			func(
				rule string,
				namespace string,
				checked bool,
			) {
				filterFunc, err := log.ParseRule(rule)
				if err != nil {
					return
				}

				logger, _ := makeLogger(filterFunc)
				defer logger.Sync() //nolint:errcheck

				if namespace != "" {
					logger = logger.Named(namespace)
				}

				entry := logger.Check(zap.DebugLevel, "")
				if checked {
					Expect(entry).NotTo(BeNil())
				} else {
					Expect(entry).To(BeNil())
				}
			},
			Entry("1", "", "", true),
			Entry("2", "", "foo", true),
			Entry("3", "*", "", true),
			Entry("4", "*", "foo", true),
			Entry("5", "*:foo", "", false),
			Entry("6", "*:foo", "foo", true),
			Entry("7", "*:foo", "bar", false),
		)

		DescribeTable("any level",
			func(name string, expected bool) {
				logger, _ := makeLogger(mustParseRules("debug:*.*", "info:demo*"))
				if name != "" {
					logger = logger.Named(name)
				}

				Expect(checkAnyLevel(logger)).To(Equal(expected))
			},
			Entry("1", "", false),
			Entry("2", "demo", true),
			Entry("3", "blahdemo", false),
			Entry("4", "demoblah", true),
			Entry("5", "blah", false),
			Entry("6", "blah.blah", true),
		)

		DescribeTable("level",
			func(name string, lvl log.Level, expected bool) {
				logger, _ := makeLogger(mustParseRules("=debug:*.*", "=info:demo*"))

				if name != "" {
					logger = logger.Named(name)
				}

				Expect(checkLevel(logger, lvl)).To(Equal(expected))
			},
			Entry("1", "", log.DebugLevel, false),
			Entry("2", "demo", log.DebugLevel, false),
			Entry("3", "blahdemo", log.DebugLevel, false),
			Entry("4", "demoblah", log.DebugLevel, false),
			Entry("5", "blah", log.DebugLevel, false),
			Entry("6", "blah.blah", log.DebugLevel, true),
			Entry("7", "", log.InfoLevel, false),
			Entry("8", "demo", log.InfoLevel, true),
			Entry("9", "blahdemo", log.InfoLevel, false),
			Entry("10", "demoblah", log.InfoLevel, true),
			Entry("11", "blah", log.InfoLevel, false),
			Entry("12", "blah.blah", log.InfoLevel, false),
		)
	})

	It("With", func() {
		logger, logs := makeLogger(log.ByNamespaces("demo1.*,demo3.*"))
		defer logger.Sync() //nolint:errcheck

		logger.With(zap.String("lorem", "ipsum")).Debug("hello city!")
		Expect(logs.TakeAll()).To(BeEmpty())

		logger.With(zap.String("lorem", "ipsum")).Named("demo1.frontend").Debug("hello region!")
		Expect(logs.TakeAll()).To(HaveExactElements(
			matchEntry(zapcore.Entry{Message: "hello region!", LoggerName: "demo1.frontend", Level: zapcore.DebugLevel}, zap.String("lorem", "ipsum")),
		))

		logger.With(zap.String("lorem", "ipsum")).Named("demo2.frontend").Debug("hello planet!")
		Expect(logs.TakeAll()).To(BeEmpty())

		logger.With(zap.String("lorem", "ipsum")).Named("demo3.frontend").Debug("hello solar system!")
		Expect(logs.TakeAll()).To(HaveExactElements(
			matchEntry(zapcore.Entry{Message: "hello solar system!", LoggerName: "demo3.frontend", Level: zapcore.DebugLevel}, zap.String("lorem", "ipsum")),
		))
	})

	It("Check", func() {
		logger, logs := makeLogger(mustParseRules("=debug:*", "=info:demo*"))
		defer logger.Sync() //nolint:errcheck

		ce := logger.Check(zap.DebugLevel, "a")
		Expect(ce).NotTo(BeNil())
		ce.Write()
		Expect(logs.TakeAll()).To(HaveExactElements(
			matchEntry(zapcore.Entry{Message: "a"}),
		))

		ce = logger.Check(zap.InfoLevel, "b")
		Expect(ce).To(BeNil())
		Expect(logs.TakeAll()).To(BeEmpty())

		ce = logger.Named("demo").Check(zap.InfoLevel, "c")
		Expect(ce).NotTo(BeNil())
		ce.Write()
		Expect(logs.TakeAll()).To(HaveExactElements(
			matchEntry(zapcore.Entry{Message: "c"}),
		))

		ce = logger.Check(zap.WarnLevel, "d")
		Expect(ce).To(BeNil())
		Expect(logs.TakeAll()).To(BeEmpty())
	})
})
