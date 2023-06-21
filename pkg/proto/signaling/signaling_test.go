// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package signaling_test

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/stv0g/cunicu/pkg/crypto"
	epdiscproto "github.com/stv0g/cunicu/pkg/proto/feature/epdisc"
	signalingproto "github.com/stv0g/cunicu/pkg/proto/signaling"
	"github.com/stv0g/cunicu/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	test.SetupLogging()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Protobuf Suite")
}

var _ = Describe("message encryption", func() {
	var c epdiscproto.Candidate
	var ourKP, theirKP *crypto.KeyPair
	var em signalingproto.EncryptedMessage

	BeforeEach(func() {
		var err error

		c = epdiscproto.Candidate{
			Foundation: "1234",
		}

		ourKP, theirKP, err = test.GenerateKeyPairs()
		Expect(err).To(Succeed())

		em = signalingproto.EncryptedMessage{}
		err = em.Marshal(&c, ourKP)
		Expect(err).To(Succeed(), "Failed to encrypt message: %s", err)
	})

	It("can en/decrypt a message", func() {
		c2 := epdiscproto.Candidate{}
		err := em.Unmarshal(&c2, theirKP)

		Expect(err).To(Succeed(), "Failed to decrypt message: %s", err)
		Expect(proto.Equal(&c, &c2)).To(BeTrue())
	})

	It("fails to decrypt an altered message", func() {
		em.Body[0] ^= 1

		c2 := epdiscproto.Candidate{}
		err := em.Unmarshal(&c2, theirKP)

		Expect(err).To(HaveOccurred(), "Decrypted invalid message: %s", err)
	})
})
