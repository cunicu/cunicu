package config_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stv0g/cunicu/pkg/config"
)

var _ = Context("default", func() {
	It("has a default hostname", func() {
		err := config.InitDefaults()
		Expect(err).To(Succeed())

		Expect(config.DefaultInterfaceSettings.HostName).NotTo(BeEmpty())
	})
})
