package crypto_test

import (
	"encoding/json"
	"net"

	"riasc.eu/wice/pkg/crypto"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PBKDF2 Key derivation", func() {
	var key1, key2 crypto.Key

	BeforeEach(func() {
		key1 = crypto.GenerateKeyFromPassword("test")
		key2 = crypto.GenerateKeyFromPassword("test2")
	})

	It("matches well known key", func() {
		Expect(crypto.ParseKey("SAyMLIWTO+DSnTx/JDak+lRR5huci8m4JsEabkkIxFY=")).To(Equal(key1))
	})

	It("does not create equal keys", func() {
		Expect(key1).NotTo(Equal(key2))
	})

	It("produces correct key length", func() {
		Expect(key1).To(HaveLen(crypto.KeyLength))
	})

	It("does not produce empty keys", func() {
		Expect(key1.IsSet()).To(BeTrue())
	})

	It("can parse the generated key as a valid key", func() {
		_, err := crypto.ParseKeyBytes(key1[:])
		Expect(err).To(Succeed())
	})
})

var _ = Describe("Key generation", func() {
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

	It("does not produce empty keys", func() {
		Expect(key1.IsSet()).To(BeTrue())
	})

	It("can parse the generated key as a valid key", func() {
		_, err := crypto.ParseKeyBytes(key1[:])
		Expect(err).To(Succeed())
	})
})

var _ = Describe("Private key generation", func() {
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

	It("does not produce empty keys", func() {
		Expect(key1.IsSet()).To(BeTrue())
	})

	It("can parse the generated key as a valid key", func() {
		_, err := crypto.ParseKeyBytes(key1[:])
		Expect(err).To(Succeed())
	})
})

var _ = Describe("Key to string conversion", Ordered, func() {
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

var _ = Describe("Key to byte slice conversions", Ordered, func() {
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

var _ = It("can marshal a Key to a string via the encoding.TextMarshaler interface", func() {
	var err error
	var obj1, obj2 struct {
		Key crypto.Key
	}

	obj1.Key, err = crypto.GeneratePrivateKey()
	Expect(err).To(Succeed())

	objJSON, err := json.Marshal(&obj1)
	Expect(err).To(Succeed())

	err = json.Unmarshal(objJSON, &obj2)
	Expect(err).To(Succeed())

	Expect(obj1).To(Equal(obj2))
})

var _ = It("can derive a public from a private key", func() {
	sk, err := crypto.ParseKey("GMHOtIxfUrGmncORjYK/slCSK/8V2TF9MjzzoPDTkEc=")
	Expect(err).To(Succeed())

	pk, err := crypto.ParseKey("Hxm0/KTFRGFirpOoTWO2iMde/gJX+oVswUXEzVN5En8=")
	Expect(err).To(Succeed())

	Expect(sk.PublicKey()).To(Equal(pk))
})

var _ = Describe("Key test", func() {
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

var _ = It("can derive a shared key via the X25519 Diffie Hellman function", func() {
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
})

var _ = It("can generate valid IPv6 link-local addresses from a public key", func() {
	key, err := crypto.GeneratePrivateKey()
	Expect(err).To(Succeed())

	addr := key.PublicKey().IPv6Address()

	_, ll, err := net.ParseCIDR("fe80::/10")
	Expect(err).To(Succeed())

	ones, bits := addr.Mask.Size()
	Expect(ones).To(Equal(64))
	Expect(bits).To(Equal(net.IPv6len * 8))
	Expect(ll.Contains(addr.IP)).To(BeTrue())
})
