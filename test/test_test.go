// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package test_test

import (
	"crypto/rand"
	"testing"

	"github.com/stv0g/cunicu/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	test.SetupLogging()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Helper Suite")
}

var _ = Describe("entropy", func() {
	Specify("that the entropy of an empty slice is zero", func() {
		Expect(test.Entropy([]byte{})).To(BeNumerically("==", 0))
		Expect(test.Entropy(nil)).To(BeNumerically("==", 0))
	})

	Specify("that the entropy of A is defined as", func() {
		Expect(test.Entropy([]byte("AAAAAAAAAAAAAA"))).To(BeZero())
	})

	Specify("that the entropy of A is defined as", func() {
		Expect(test.Entropy([]byte("This is some not-so random string"))).To(
			And(
				BeNumerically(">", 1),
				BeNumerically("<", 5),
			),
		)
	})

	Specify("that the entropy of random data", func() {
		random := make([]byte, 128)
		n, err := rand.Read(random)
		Expect(err).To(Succeed())
		Expect(n).To(Equal(128))

		Expect(test.Entropy(random)).To(BeNumerically(">", 5))
	})
})
