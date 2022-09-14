package config_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stv0g/cunicu/pkg/config"
)

var _ = Context("meta", func() {
	var m *config.Meta

	BeforeEach(func() {
		m = config.Metadata()
		Expect(m).NotTo(BeNil())
	})

	It("can enumerate all keys", func() {
		m := config.Metadata()
		keys := m.Keys()

		Expect(keys).To(ContainElements(
			Equal("hsync.enabled"),
			Equal("watch_interval"),
			Equal("epdisc.ice.disconnected_timeout"),
		))

		Expect(keys).NotTo(ContainElements(
			"epdisc.ice",
		))
	})

	It("can lookup a key", func() {
		n := m.Lookup("epdisc.ice")
		Expect(n.Fields).To(HaveKey("disconnected_timeout"))
	})
})
