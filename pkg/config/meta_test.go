// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"github.com/stv0g/cunicu/pkg/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
			Equal("sync_hosts"),
			Equal("watch_interval"),
			Equal("ice.disconnected_timeout"),
		))

		Expect(keys).NotTo(ContainElements(
			"ice",
		))
	})

	It("can lookup a key", func() {
		n := m.Lookup("ice")
		Expect(n.Fields).To(HaveKey("disconnected_timeout"))
	})
})
