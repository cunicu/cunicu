// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package tty_test

import (
	"bytes"
	"fmt"

	"github.com/stv0g/cunicu/pkg/tty"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("Indenter", func() {
	It("can indent", func() {
		b := bytes.NewBuffer(nil)
		i := tty.NewIndenter(b, "--")

		n, err := fmt.Fprint(i, "Hello World")
		Expect(err).To(Succeed())
		Expect(n).To(Equal(13))
		Expect(b.String()).To(Equal("--Hello World"))

		n, err = fmt.Fprint(i, "\nThis")
		Expect(err).To(Succeed())
		Expect(n).To(Equal(7))
		Expect(b.String()).To(Equal("--Hello World\n--This"))

		n, err = fmt.Fprint(i, " is indented\n")
		Expect(err).To(Succeed())
		Expect(n).To(Equal(13))
		Expect(b.String()).To(Equal("--Hello World\n--This is indented\n"))

		n, err = fmt.Fprint(i, "Good night")
		Expect(err).To(Succeed())
		Expect(n).To(Equal(12))
		Expect(b.String()).To(Equal("--Hello World\n--This is indented\n--Good night"))
	})
})
