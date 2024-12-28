// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package os_test

import (
	"time"

	osx "cunicu.li/cunicu/pkg/os"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("Clock", func() {
	It("does not error", func() {
		_, err := osx.GetClockMonotonic()
		Expect(err).To(Succeed())
	})

	It("does not goes backwards", func() {
		t1, err := osx.GetClockMonotonic()
		Expect(err).To(Succeed())

		time.Sleep(10 * time.Millisecond)

		t2, err := osx.GetClockMonotonic()
		Expect(err).To(Succeed())

		Expect(t1.Before(t2)).To(BeTrue())
	})
})
