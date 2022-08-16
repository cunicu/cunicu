package pb_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/test"
)

var _ = Describe("message encryption", func() {
	var c pb.Candidate
	var ourKP, theirKP *crypto.KeyPair
	var em pb.EncryptedMessage

	BeforeEach(func() {
		var err error

		c = pb.Candidate{
			Foundation: "1234",
		}

		ourKP, theirKP, err = test.GenerateKeyPairs()
		Expect(err).To(Succeed())

		em = pb.EncryptedMessage{}
		err = em.Marshal(&c, ourKP)
		Expect(err).To(Succeed(), "Failed to encrypt message: %s", err)
	})

	It("can en/decrypt a message", func() {
		c2 := pb.Candidate{}
		err := em.Unmarshal(&c2, theirKP)

		Expect(err).To(Succeed(), "Failed to decrypt message: %s", err)
		Expect(proto.Equal(&c, &c2)).To(BeTrue())
	})

	It("fails to decrypt an altered message", func() {
		em.Body[0] ^= 1

		c2 := pb.Candidate{}
		err := em.Unmarshal(&c2, theirKP)

		Expect(err).To(HaveOccurred(), "Decrypted invalid message: %s", err)
	})
})
