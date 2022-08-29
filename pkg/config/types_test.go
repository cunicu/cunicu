package config_test

import (
	"net/url"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
	"riasc.eu/wice/pkg/config"
)

var _ = Context("types", func() {
	Context("regex", func() {
		const re1Str = "[a-z]"
		const re2Str = "[a-z"
		var re1 = regexp.MustCompile(re1Str)

		It("unmarshal", func() {
			var re config.Regexp
			err := re.UnmarshalText([]byte(re1Str))

			Expect(err).To(Succeed())

			Expect(re.MatchString("1")).NotTo(BeTrue())
			Expect(re.MatchString("a")).To(BeTrue())
		})

		It("marshal", func() {
			re := config.Regexp{*re1}

			reStr, err := re.MarshalText()
			Expect(err).To(Succeed())
			Expect(string(reStr)).To(Equal(re1Str), "MarshalText: %s != %s", string(reStr), re1Str)
		})

		It("fails on invalid regex", func() {
			var re config.Regexp
			err := re.UnmarshalText([]byte(re2Str))
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
	})

})
