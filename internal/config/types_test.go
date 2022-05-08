package config_test

import (
	"net/url"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"riasc.eu/wice/internal/config"
)

var _ = Context("regex", func() {
	const re1Str = "[a-z]"
	var re1 = regexp.MustCompile(re1Str)

	It("Unmarshal", func() {
		var re config.Regexp
		err := re.UnmarshalText([]byte(re1Str))

		Expect(err).To(Succeed())

		Expect(re.MatchString("1")).NotTo(BeTrue())
		Expect(re.MatchString("a")).To(BeTrue())
	})

	It("Marshal", func() {
		re := config.Regexp{*re1}

		reStr, err := re.MarshalText()
		Expect(err).To(Succeed())
		Expect(string(reStr)).To(Equal(re1Str), "MarshalText: %s != %s", string(reStr), re1Str)
	})
})

var _ = Context("backend url", func() {
	const urlStr = "https://example.com"
	var urlExp = config.BackendURL{
		URL: url.URL{
			Scheme: "https",
			Host:   "example.com",
		},
	}

	It("Unmarshal", func() {
		var u config.BackendURL
		err := u.UnmarshalText([]byte(urlStr))

		Expect(err).To(Succeed())
		Expect(u).To(Equal(urlExp))
	})

	It("Marshal", func() {
		u, err := urlExp.MarshalText()
		Expect(err).To(Succeed())
		Expect(u).To(BeEquivalentTo(urlStr))
	})
})
