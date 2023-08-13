// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package crypto_test

import (
	"testing"

	"cunicu.li/cunicu/pkg/crypto"
	"cunicu.li/cunicu/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	test.SetupLogging()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Crypto Suite")
}

var _ = Describe("nonce", func() {
	It("can generate a valid nonce", func() {
		nonce, err := crypto.GetNonce(100)
		Expect(err).To(Succeed())
		Expect(nonce).To(HaveLen(100))
		Expect([]byte(nonce)).To(test.BeRandom())
	})
})
