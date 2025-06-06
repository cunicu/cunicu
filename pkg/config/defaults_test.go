// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"cunicu.li/cunicu/pkg/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("default", func() {
	It("has a default hostname", func() {
		err := config.InitDefaults()
		Expect(err).To(Succeed())

		Expect(config.DefaultSettings.DefaultInterfaceSettings.HostName).NotTo(BeEmpty())
	})
})
