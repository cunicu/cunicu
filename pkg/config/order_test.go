// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"github.com/stv0g/cunicu/pkg/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = It("Can extract interface order", func() {
	cfg := `---
interfaces:
  f:
  c:
  b:
  a:
  d:
  e:
`

	order, err := config.ExtractInterfaceOrder([]byte(cfg))
	Expect(err).To(Succeed())
	Expect(order).To(Equal([]string{"f", "c", "b", "a", "d", "e"}))
})
