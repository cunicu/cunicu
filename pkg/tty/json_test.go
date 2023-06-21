// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package tty_test

import (
	"github.com/stv0g/cunicu/pkg/tty"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("re-indent json", func() {
	It("works", func() {
		original := []byte(`{ "a": { "b": { "c": 5 } } }`)
		indented := []byte(`{
  "a": {
    "b": {
      "c": 5
    }
  }
}`)

		Expect(tty.ReIndentJSON(original, "", "  ")).To(Equal(indented))
	})

	It("fails for invalid json", func() {
		_, err := tty.ReIndentJSON([]byte("{"), "", "  ")
		Expect(err).To(HaveOccurred())
	})
})
