// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-FileCopyrightText: 2020 Manfred Touron <https://manfred.life>
// SPDX-License-Identifier: Apache-2.0

package log_test

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gcustom"
	"github.com/stv0g/cunicu/pkg/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

var (
	errMismatchingLogMessage = errors.New("mismatch in log message")
	errMismatchingLoggerName = errors.New("mismatch in logger name")
	errMismatchingLogLevel   = errors.New("mismatch in log level")
	errMismatchingField      = errors.New("mismatch in context field")
)

func MatchEntry(expectedEntry zapcore.Entry, expectedFields ...zapcore.Field) OmegaMatcher {
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

	rule := new(log.AtomicFilterRule)
	rule.Store(log.NewFilterRule(filterFunc))

	filtered := log.NewFilteredCore(observed, rule)

	return zap.New(filtered), logs
}

var _ = Describe("filter", func() {
	Describe("NewFilteredCore", func() {
		It("wrap", func() {
			next, logs := observer.New(zapcore.DebugLevel)
			logger := zap.New(next)
			defer logger.Sync() //nolint:errcheck

			rule := new(log.AtomicFilterRule)
			rule.Store(log.NewFilterRule(log.MustParseRules("demo*")))

			logger = logger.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
				return log.NewFilteredCore(c, rule)
			}))

			logger.Debug("hello world!")
			logger.Named("demo").Debug("hello earth!")
			logger.Named("other").Debug("hello universe!")

			Expect(logs.All()).To(HaveExactElements(
				MatchEntry(zapcore.Entry{Message: "hello earth!"}),
			))
		})

		It("new logger", func() {
			logger, logs := makeLogger(log.MustParseRules("demo*"))
			defer logger.Sync() //nolint:errcheck

			logger.Debug("hello world!")
			logger.Named("demo").Debug("hello earth!")
			logger.Named("other").Debug("hello universe!")

			Expect(logs.All()).To(HaveExactElements(
				MatchEntry(zapcore.Entry{Message: "hello earth!"}),
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
				MatchEntry(zapcore.Entry{Message: "hello region!", LoggerName: "demo1.frontend", Level: zapcore.DebugLevel}),
				MatchEntry(zapcore.Entry{Message: "hello solar system!", LoggerName: "demo3.frontend", Level: zapcore.DebugLevel}),
			))
		})

		It("custom", func() {
			rand.Seed(42) //nolint:staticcheck

			logger, logs := makeLogger(func(entry zapcore.Entry, fields []zapcore.Field) bool {
				return rand.Intn(2) == 1 //nolint:gosec
			})
			defer logger.Sync() //nolint:errcheck

			logger.Debug("hello city!")
			logger.Debug("hello region!")
			logger.Debug("hello planet!")
			logger.Debug("hello solar system!")
			logger.Debug("hello universe!")
			logger.Debug("hello multiverse!")

			Expect(logs.All()).To(HaveExactElements(
				MatchEntry(zapcore.Entry{Message: "hello city!"}),
				MatchEntry(zapcore.Entry{Message: "hello solar system!"}),
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
				func(entry zapcore.Entry, fields []zapcore.Field) bool {
					return true
				},
				[]string{"a", "b", "c", "d"},
			),
			Entry("disallow-all",
				func(entry zapcore.Entry, fields []zapcore.Field) bool {
					return false
				},
				[]string{},
			),
			Entry("minimum-debug",
				log.MinimumLevel(zapcore.DebugLevel),
				[]string{"a", "b", "c", "d"},
			),
			Entry("minimum-info",
				log.MinimumLevel(zapcore.InfoLevel),
				[]string{"b", "c", "d"},
			),
			Entry("minimum-warn",
				log.MinimumLevel(zapcore.WarnLevel),
				[]string{"c", "d"},
			),
			Entry("minimum-error",
				log.MinimumLevel(zapcore.ErrorLevel),
				[]string{"d"},
			),
			Entry("exact-debug",
				log.ExactLevel(zapcore.DebugLevel),
				[]string{"a"},
			),
			Entry("exact-info",
				log.ExactLevel(zapcore.InfoLevel),
				[]string{"b"},
			),
			Entry("exact-warn",
				log.ExactLevel(zapcore.WarnLevel),
				[]string{"c"},
			),
			Entry("exact-error",
				log.ExactLevel(zapcore.ErrorLevel),
				[]string{"d"},
			),
			Entry("all-except-debug",
				log.Reverse(log.ExactLevel(zapcore.DebugLevel)),
				[]string{"b", "c", "d"},
			),
			Entry("all-except-info",
				log.Reverse(log.ExactLevel(zapcore.InfoLevel)),
				[]string{"a", "c", "d"},
			),
			Entry("all-except-warn",
				log.Reverse(log.ExactLevel(zapcore.WarnLevel)),
				[]string{"a", "b", "d"},
			),
			Entry("all-except-error",
				log.Reverse(log.ExactLevel(zapcore.ErrorLevel)),
				[]string{"a", "b", "c"},
			),
			Entry("any",
				log.Any(
					log.ExactLevel(zapcore.DebugLevel),
					log.ExactLevel(zapcore.WarnLevel),
				),
				[]string{"a", "c"},
			),
			Entry("all-1",
				log.All(
					log.ExactLevel(zapcore.DebugLevel),
					log.ExactLevel(zapcore.WarnLevel),
				),
				[]string{},
			),
			Entry("all-2",
				log.All(
					log.ExactLevel(zapcore.DebugLevel),
					log.ExactLevel(zapcore.DebugLevel),
				),
				[]string{"a"},
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
		// *=myns             => any level, myns namespace
		// info,warn:myns.*   => info or warn level, any namespace matching myns.*
		// error=*            => everything with error level
		logger, logs := makeLogger(log.MustParseRules("*:myns info,warn:myns.* error:*"))
		defer logger.Sync() //nolint:errcheck

		logger.Debug("top debug") // No match
		Expect(logs.TakeAll()).To(BeEmpty())

		logger.Named("myns").Debug("myns debug") // Matches *:myns
		Expect(logs.TakeAll()).To(HaveExactElements(
			MatchEntry(zapcore.Entry{Message: "myns debug", LoggerName: "myns", Level: zapcore.DebugLevel}),
		))

		logger.Named("bar").Debug("bar debug") // No match
		Expect(logs.TakeAll()).To(BeEmpty())

		logger.Named("myns").Named("foo").Debug("myns.foo debug") // No match
		Expect(logs.TakeAll()).To(BeEmpty())

		logger.Info("top info") // No match
		Expect(logs.TakeAll()).To(BeEmpty())

		logger.Named("myns").Info("myns info") // Matches *:myns
		Expect(logs.TakeAll()).To(HaveExactElements(
			MatchEntry(zapcore.Entry{Message: "myns info", LoggerName: "myns", Level: zapcore.InfoLevel}),
		))

		logger.Named("bar").Info("bar info") // No match
		Expect(logs.TakeAll()).To(BeEmpty())

		logger.Named("myns").Named("foo").Info("myns.foo info") // Matches info,warn:myns.*
		Expect(logs.TakeAll()).To(HaveExactElements(
			MatchEntry(zapcore.Entry{Message: "myns.foo info", LoggerName: "myns.foo", Level: zapcore.InfoLevel}),
		))

		logger.Warn("top warn") // No match
		Expect(logs.TakeAll()).To(BeEmpty())

		logger.Named("myns").Warn("myns warn") // Matches *:myns
		Expect(logs.TakeAll()).To(HaveExactElements(
			MatchEntry(zapcore.Entry{Message: "myns warn", LoggerName: "myns", Level: zapcore.WarnLevel}),
		))

		logger.Named("bar").Warn("bar warn") // No match
		Expect(logs.TakeAll()).To(BeEmpty())

		logger.Named("myns").Named("foo").Warn("myns.foo warn") // Matches info,warn:myns.*
		Expect(logs.TakeAll()).To(HaveExactElements(
			MatchEntry(zapcore.Entry{Message: "myns.foo warn", LoggerName: "myns.foo", Level: zapcore.WarnLevel}),
		))

		logger.Error("top error") // Matches error:*
		Expect(logs.TakeAll()).To(HaveExactElements(
			MatchEntry(zapcore.Entry{Message: "top error", Level: zapcore.ErrorLevel}),
		))

		logger.Named("myns").Error("myns error") // Matches *:myns and error:*
		Expect(logs.TakeAll()).To(HaveExactElements(
			MatchEntry(zapcore.Entry{Message: "myns error", LoggerName: "myns", Level: zapcore.ErrorLevel}),
		))

		logger.Named("bar").Error("bar error") // Matches error:*
		Expect(logs.TakeAll()).To(HaveExactElements(
			MatchEntry(zapcore.Entry{Message: "bar error", LoggerName: "bar", Level: zapcore.ErrorLevel}),
		))

		logger.Named("myns").Named("foo").Error("myns.foo error") // Matches error:*
		Expect(logs.TakeAll()).To(HaveExactElements(
			MatchEntry(zapcore.Entry{Message: "myns.foo error", LoggerName: "myns.foo", Level: zapcore.ErrorLevel}),
		))
	})

	DescribeTable("ParseRules",
		func(
			input string,
			expectedLogs string,
			expectedError error,
		) {
			filterFunc, err := log.ParseRules(input)
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
		Entry("empty", "", "", nil),
		Entry("everything", "*", everything, nil),
		Entry("debug+", "debug+:*", everything, nil),
		Entry("all-debug", "debug:*", allDebug, nil),
		Entry("all-info", "info:*", allInfo, nil),
		Entry("all-warn", "warn:*", allWarn, nil),
		Entry("all-error", "error:*", allError, nil),
		Entry("all-info-and-warn-1", "info,warn:*", "bcfgjknorsvwz034", nil),
		Entry("all-info-and-warn-2", "info:* warn:*", "bcfgjknorsvwz034", nil),
		Entry("warn+", "warn+:*", "cdghklopstwx0145", nil),
		Entry("redundant-1", "info,info:* info:*", allInfo, nil),
		Entry("redundant-2", "* *:* info:*", everything, nil),
		Entry("foo-ns", "foo", "efgh", nil),
		Entry("foo-ns-wildcard", "*:foo", "efgh", nil),
		Entry("foo-ns-debug,info", "debug,info:foo", "ef", nil),
		Entry("foo.star-ns", "foo.*", "qrstuvwx", nil),
		Entry("foo.star-ns-wildcard", "*:foo.*", "qrstuvwx", nil),
		Entry("foo.star-ns-debug,info", "debug,info:foo.*", "qruv", nil),
		Entry("all-in-one", "*:foo debug:foo.* info,warn:bar error:*", "defghjklpqtux15", nil),
		Entry("exclude-1", "info:test,foo*,-foo.foo", "fr", nil),
		Entry("exclude-2", "info:test,foo*,-*.foo", "fr", nil),
		Entry("exclude-3", "test,*.foo,-foo.*", "yz012345", nil),
		Entry("exclude-4", "*,-foo,-bar", "abcdmnopqrstuvwxyz012345", nil),
		Entry("exclude-5", "foo*,bar*,-foo.foo,-bar.foo", "efghijklqrst", nil),
		Entry("exclude-6", "foo*,-foo.foo,bar*,-bar.foo", "efghijklqrst", nil),
		Entry("invalid-left", "invalid:*", "", log.ErrUnsupportedKeyword),
		Entry("missing-left", ":*", "", log.ErrBadSyntax),
		Entry("missing-right", ":*", "", log.ErrBadSyntax),
		PEntry("missing-exclude-pattern", "*:-", "", log.ErrBadSyntax),
	)

	Describe("Check", func() {
		DescribeTable("simple",
			func(
				rules string,
				namespace string,
				checked bool,
			) {
				filterFunc, err := log.ParseRules(rules)
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
			Entry(nil, "", "", false),
			Entry(nil, "", "foo", false),
			Entry(nil, "*", "", true),
			Entry(nil, "*", "foo", true),
			Entry(nil, "*:foo", "", false),
			Entry(nil, "*:foo", "foo", true),
			Entry(nil, "*:foo", "bar", false),
		)

		DescribeTable("any level",
			func(name string, expected bool) {
				logger, _ := makeLogger(log.MustParseRules("debug:*.* info:demo*"))
				if name != "" {
					logger = logger.Named(name)
				}

				Expect(log.CheckAnyLevel(logger)).To(Equal(expected))
			},
			Entry(nil, "", false),
			Entry(nil, "demo", true),
			Entry(nil, "blahdemo", false),
			Entry(nil, "demoblah", true),
			Entry(nil, "blah", false),
			Entry(nil, "blah.blah", true),
		)

		DescribeTable("level",
			func(name string, lvl zapcore.Level, expected bool) {
				logger, _ := makeLogger(log.MustParseRules("debug:*.* info:demo*"))

				if name != "" {
					logger = logger.Named(name)
				}

				Expect(log.CheckLevel(logger, lvl)).To(Equal(expected))
			},
			Entry(nil, "", zap.DebugLevel, false),
			Entry(nil, "demo", zap.DebugLevel, false),
			Entry(nil, "blahdemo", zap.DebugLevel, false),
			Entry(nil, "demoblah", zap.DebugLevel, false),
			Entry(nil, "blah", zap.DebugLevel, false),
			Entry(nil, "blah.blah", zap.DebugLevel, true),
			Entry(nil, "", zap.InfoLevel, false),
			Entry(nil, "demo", zap.InfoLevel, true),
			Entry(nil, "blahdemo", zap.InfoLevel, false),
			Entry(nil, "demoblah", zap.InfoLevel, true),
			Entry(nil, "blah", zap.InfoLevel, false),
			Entry(nil, "blah.blah", zap.InfoLevel, false),
		)
	})

	It("With", func() {
		logger, logs := makeLogger(log.ByNamespaces("demo1.*,demo3.*"))
		defer logger.Sync() //nolint:errcheck

		logger.With(zap.String("lorem", "ipsum")).Debug("hello city!")
		Expect(logs.TakeAll()).To(BeEmpty())

		logger.With(zap.String("lorem", "ipsum")).Named("demo1.frontend").Debug("hello region!")
		Expect(logs.TakeAll()).To(HaveExactElements(
			MatchEntry(zapcore.Entry{Message: "hello region!", LoggerName: "demo1.frontend", Level: zapcore.DebugLevel}, zap.String("lorem", "ipsum")),
		))

		logger.With(zap.String("lorem", "ipsum")).Named("demo2.frontend").Debug("hello planet!")
		Expect(logs.TakeAll()).To(BeEmpty())

		logger.With(zap.String("lorem", "ipsum")).Named("demo3.frontend").Debug("hello solar system!")
		Expect(logs.TakeAll()).To(HaveExactElements(
			MatchEntry(zapcore.Entry{Message: "hello solar system!", LoggerName: "demo3.frontend", Level: zapcore.DebugLevel}, zap.String("lorem", "ipsum")),
		))
	})

	It("Check", func() {
		logger, logs := makeLogger(log.MustParseRules("debug:* info:demo*"))
		defer logger.Sync() //nolint:errcheck

		ce := logger.Check(zap.DebugLevel, "a")
		Expect(ce).NotTo(BeNil())
		ce.Write()
		Expect(logs.TakeAll()).To(HaveExactElements(
			MatchEntry(zapcore.Entry{Message: "a"}),
		))

		ce = logger.Check(zap.InfoLevel, "b")
		Expect(ce).To(BeNil())
		Expect(logs.TakeAll()).To(BeEmpty())

		ce = logger.Named("demo").Check(zap.InfoLevel, "c")
		Expect(ce).NotTo(BeNil())
		ce.Write()
		Expect(logs.TakeAll()).To(HaveExactElements(
			MatchEntry(zapcore.Entry{Message: "c"}),
		))

		ce = logger.Check(zap.WarnLevel, "d")
		Expect(ce).To(BeNil())
		Expect(logs.TakeAll()).To(BeEmpty())
	})
})
