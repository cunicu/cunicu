// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package tty_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stv0g/cunicu/pkg/tty"
	"github.com/stv0g/cunicu/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	test.SetupLogging()
	RegisterFailHandler(Fail)
	RunSpecs(t, "TTY Suite")
}

var _ = Context("IsATTY", func() {
	if test.IsCI() {
		It("is false in CI runners", func() {
			Expect(tty.IsATTY(os.Stdout)).To(BeFalse())
		})
	} else {
		It("is true outside CI runners", func() {
			Expect(tty.IsATTY(os.Stdout)).To(BeTrue())
		})
	}

	It("is false on files", func() {
		fn := filepath.Join(GinkgoT().TempDir(), "file")
		f, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY, 0o600)
		Expect(err).To(Succeed())

		Expect(tty.IsATTY(f)).To(BeFalse())

		err = f.Close()
		Expect(err).To(Succeed())
	})
})
