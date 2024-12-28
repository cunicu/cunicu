// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package crypto_test

import (
	"testing"

	"cunicu.li/cunicu/pkg/crypto"
	"cunicu.li/cunicu/pkg/tty"
	"cunicu.li/cunicu/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	test.SetupLogging()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Crypto Suite")
}

var _ = Describe("generate", func() {
	It("nonce", func() {
		nonce, err := crypto.GetNonce(100)
		Expect(err).To(Succeed())
		Expect(nonce).To(HaveLen(100))
		Expect([]byte(nonce)).To(test.BeRandom())
	})

	It("random string", func() {
		for range 10000 {
			s, err := crypto.GetRandomString(10, tty.RunesAlpha)
			Expect(err).To(Succeed())
			Expect(s).To(HaveLen(10))
			Expect(s).To(MatchRegexp(`^[a-zA-Z]+$`), "Generator returned unexpected character")
		}
	})
})
