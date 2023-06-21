// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package crypto_test

import (
	"net"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("key", func() {
	Describe("Argon2id key derivation", func() {
		var key1, key2 crypto.Key

		BeforeEach(func() {
			key1 = crypto.GenerateKeyFromPassword("test")
			key2 = crypto.GenerateKeyFromPassword("test2")
		})

		It("matches well known key", func() {
			Expect(crypto.ParseKey("KJJj36cAiOLIaAImbnZtzvk6KmIpx87LLC4sCnriuUw=")).To(Equal(key1))
		})

		It("does not create equal keys", func() {
			Expect(key1).NotTo(Equal(key2))
		})

		It("produces correct key length", func() {
			Expect(key1).To(HaveLen(crypto.KeyLength))
		})

		It("produces a random key", func() {
			Expect(key1.Bytes()).To(test.BeRandom())
		})

		It("can parse the generated key as a valid key", func() {
			_, err := crypto.ParseKeyBytes(key1[:])
			Expect(err).To(Succeed())
		})
	})

	Describe("generation", func() {
		var key1, key2 crypto.Key

		BeforeEach(func() {
			var err error
			key1, err = crypto.GenerateKey()
			Expect(err).To(Succeed())

			key2, err = crypto.GenerateKey()
			Expect(err).To(Succeed())
		})

		It("does not create equal keys", func() {
			Expect(key1).NotTo(Equal(key2))
		})

		It("produces correct key length", func() {
			Expect(key1).To(HaveLen(crypto.KeyLength))
		})

		It("generates random keys", func() {
			Expect(key1.Bytes()).To(test.BeRandom())
		})

		It("can parse the generated key as a valid key", func() {
			_, err := crypto.ParseKeyBytes(key1[:])
			Expect(err).To(Succeed())
		})
	})

	Describe("private generation", func() {
		var key1, key2 crypto.Key

		BeforeEach(func() {
			var err error
			key1, err = crypto.GeneratePrivateKey()
			Expect(err).To(Succeed())

			key2, err = crypto.GeneratePrivateKey()
			Expect(err).To(Succeed())
		})

		It("does not create equal keys", func() {
			Expect(key1).NotTo(Equal(key2))
		})

		It("produces correct key length", func() {
			Expect(key1).To(HaveLen(crypto.KeyLength))
		})

		It("generates random keys", func() {
			Expect(key1.Bytes()).To(test.BeRandom())
		})

		It("can parse the generated key as a valid key", func() {
			_, err := crypto.ParseKeyBytes(key1[:])
			Expect(err).To(Succeed())
		})
	})

	Describe("ToString", Ordered, func() {
		var key1, key2 crypto.Key
		var keyString string

		BeforeAll(func() {
			var err error

			key1, err = crypto.GeneratePrivateKey()
			Expect(err).To(Succeed())
		})

		It("can generate a string of the key", func() {
			const Base64Regex = "^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$"

			keyString = key1.String()
			Expect(keyString).To(MatchRegexp(Base64Regex))
		})

		It("can parse the generated string back as a key", func() {
			var err error

			key2, err = crypto.ParseKey(keyString)
			Expect(err).To(Succeed())
		})

		It("result in the same key as the original", func() {
			Expect(key1).To(Equal(key2))
		})
	})

	Describe("parsing", func() {
		It("fails to parse a key from a malformed string", func() {
			_, err := crypto.ParseKey("this is not a proper base64 encoded key")
			Expect(err.Error()).To(ContainSubstring("illegal base64 data"))
		})

		It("fails to parse a proper base64 string with incorrect length", func() {
			_, err := crypto.ParseKey("gHI04aIM0Nopa149R+Isnj7bI+B750p2/BMwy4tV8YEC")
			Expect(err).To(MatchError("invalid length"))
		})
	})

	Describe("to and from []byte", Ordered, func() {
		var key1, key2 crypto.Key
		var keyBytes []byte

		BeforeAll(func() {
			var err error

			key1, err = crypto.GeneratePrivateKey()
			Expect(err).To(Succeed())
		})

		It("can produce a byte slice from the key", func() {
			keyBytes = key1.Bytes()
			Expect(keyBytes).To(HaveLen(crypto.KeyLength))
		})

		It("can parse the byte slice back to a key", func() {
			var err error
			key2, err = crypto.ParseKeyBytes(keyBytes)
			Expect(err).To(Succeed())
		})

		It("matches the original key", func() {
			Expect(key1).To(Equal(key2))
		})
	})

	Describe("marshaling", func() {
		var key crypto.Key
		var keyStr, brokenKeyStr string

		BeforeEach(func() {
			var err error

			key, err = crypto.GenerateKey()
			Expect(err).To(Succeed())

			keyStr = key.String()
			brokenKeyStr = keyStr[:len(keyStr)-2]
		})

		It("unmarshal", func() {
			var keyCfg crypto.Key

			err := keyCfg.UnmarshalText([]byte(keyStr))
			Expect(err).To(Succeed())

			Expect(keyCfg).To(BeEquivalentTo(key))
		})

		It("marshal", func() {
			keyStr2, err := key.MarshalText()
			Expect(err).To(Succeed())
			Expect(string(keyStr2)).To(Equal(keyStr))
		})

		It("fails on invalid key", func() {
			var k crypto.Key
			err := k.UnmarshalText([]byte(brokenKeyStr))
			Expect(err).To(HaveOccurred())
		})
	})

	It("can derive a public from a private key", func() {
		sk, err := crypto.ParseKey("GMHOtIxfUrGmncORjYK/slCSK/8V2TF9MjzzoPDTkEc=")
		Expect(err).To(Succeed())

		pk, err := crypto.ParseKey("Hxm0/KTFRGFirpOoTWO2iMde/gJX+oVswUXEzVN5En8=")
		Expect(err).To(Succeed())

		Expect(sk.PublicKey()).To(Equal(pk))
	})

	Describe("tests", func() {
		It("non-empty", func() {
			key, err := crypto.GeneratePrivateKey()
			Expect(err).To(Succeed())
			Expect(key.IsSet()).To(BeTrue())
		})

		It("empty", func() {
			key := crypto.Key{}
			Expect(key.IsSet()).To(BeFalse())
		})
	})

	Describe("address", func() {
		It("from IPv4 prefix", func() {
			_, p, err := net.ParseCIDR("10.237.0.0/16")
			Expect(err).To(Succeed())

			k, err := crypto.GenerateKey()
			Expect(err).To(Succeed())

			q := k.IPAddress(*p)

			ones, bits := q.Mask.Size()
			Expect(ones).To(Equal(16))
			Expect(bits).To(Equal(32))
			Expect(p.Contains(q.IP)).To(BeTrue())
		})

		It("from IPv6 prefix", func() {
			_, p, err := net.ParseCIDR("fc2f:9a4d::/32")
			Expect(err).To(Succeed())

			k, err := crypto.GenerateKey()
			Expect(err).To(Succeed())

			q := k.IPAddress(*p)

			ones, bits := q.Mask.Size()
			Expect(ones).To(Equal(32))
			Expect(bits).To(Equal(128))
			Expect(p.Contains(q.IP)).To(BeTrue())
		})
	})

	Describe("pair", func() {
		It("can derive a shared key via the X25519 Diffie Hellman function", func() {
			key1, err := crypto.GeneratePrivateKey()
			Expect(err).To(Succeed())

			key2, err := crypto.GeneratePrivateKey()
			Expect(err).To(Succeed())

			kp1 := crypto.KeyPair{
				Ours:   key1,
				Theirs: key2.PublicKey(),
			}

			kp2 := crypto.KeyPair{
				Ours:   key2,
				Theirs: key1.PublicKey(),
			}

			Expect(kp1.Shared()).To(Equal(kp2.Shared()))
			Expect(kp1.Shared().Bytes()).To(test.BeRandom())
		})

		It("can generate a public key pair from a public/private one", func() {
			key1, err := crypto.GeneratePrivateKey()
			Expect(err).To(Succeed())

			key2, err := crypto.GeneratePrivateKey()
			Expect(err).To(Succeed())

			kp := crypto.KeyPair{
				Ours:   key1,
				Theirs: key2.PublicKey(),
			}

			pkp := kp.Public()

			Expect(pkp.Ours).To(Equal(kp.Ours.PublicKey()))
			Expect(pkp.Theirs).To(Equal(kp.Theirs))
		})
	})
})
