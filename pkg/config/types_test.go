// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"net/url"

	"github.com/spf13/pflag"

	"github.com/stv0g/cunicu/pkg/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("types", func() {
	Context("url", func() {
		t := []TableEntry{
			Entry("example", "https://example.com", url.URL{
				Scheme: "https",
				Host:   "example.com",
			}),
			Entry("file", "file:///log.txt", url.URL{
				Scheme: "file",
				Path:   "/log.txt",
			}),
		}

		DescribeTable("unmarshal", func(urlStr string, url url.URL) {
			var u config.BackendURL
			err := u.UnmarshalText([]byte(urlStr))

			Expect(err).To(Succeed())
			Expect(u.URL).To(Equal(url))
		}, t)

		DescribeTable("marshal", func(urlStr string, url url.URL) {
			u := config.URL{url}
			m, err := u.MarshalText()
			Expect(err).To(Succeed())
			Expect(string(m)).To(BeEquivalentTo(urlStr))
		}, t)

		It("fails for invalid urls", func() {
			var u config.BackendURL
			err := u.UnmarshalText([]byte("-"))
			Expect(err).To(HaveOccurred())
		})
	})

	Context("backend url", func() {
		urls := map[string]config.BackendURL{
			"https://example.com": {
				URL: url.URL{
					Scheme: "https",
					Host:   "example.com",
				},
			},
			"p2p": {
				URL: url.URL{
					Scheme: "p2p",
				},
			},
		}

		for urlStr, urlExp := range urls {
			Context("works for valid urls", func() {
				It("unmarshal", func() {
					var u config.BackendURL
					err := u.UnmarshalText([]byte(urlStr))

					Expect(err).To(Succeed())
					Expect(u).To(Equal(urlExp))
				})

				It("marshal", func() {
					u, err := urlExp.MarshalText()
					Expect(err).To(Succeed())
					Expect(u).To(BeEquivalentTo(urlStr))
				})
			})
		}

		It("fails for invalid urls", func() {
			var u config.BackendURL
			err := u.UnmarshalText([]byte("-"))
			Expect(err).To(HaveOccurred())
		})
	})

	Context("output format", func() {
		for _, f := range config.OutputFormats {
			It(f.String(), func() {
				var g config.OutputFormat

				flags := pflag.NewFlagSet("test", pflag.ExitOnError)
				flags.Var(&g, "format", "Output format")

				err := flags.Parse([]string{"--format", f.String()})
				Expect(err).To(Succeed())
				Expect(g).To(BeEquivalentTo(f))

				h, err := f.MarshalText()
				Expect(err).To(Succeed())
				Expect(h).To(BeEquivalentTo(f.String()))
			})
		}

		It("fails on invalid format", func() {
			var g config.OutputFormat

			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			flags.Var(&g, "format", "Output format")

			err := flags.Parse([]string{"--format", "blub"})
			Expect(err).To(MatchError("invalid argument \"blub\" for \"--format\" flag: unknown output format: blub"))
		})
	})
})
