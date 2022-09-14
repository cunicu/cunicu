package crypto_test

import (
	"net"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/test"

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

	It("produces a random key", func() {
		Expect(key1.Bytes()).To(test.BeRandom())
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

	It("generates random keys", func() {
		Expect(key1.Bytes()).To(test.BeRandom())
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

	It("generates random keys", func() {
		Expect(key1.Bytes()).To(test.BeRandom())
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

var _ = Describe("Key parsing", func() {
	It("fails to parse a key from a malformed string", func() {
		_, err := crypto.ParseKey("this is not a proper base64 encoded key")
		Expect(err.Error()).To(ContainSubstring("illegal base64 data"))
	})

	It("fails to parse a proper base64 string with incorrect length", func() {
		_, err := crypto.ParseKey("gHI04aIM0Nopa149R+Isnj7bI+B750p2/BMwy4tV8YEC")
		Expect(err).To(MatchError("invalid length"))
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

var _ = Describe("key marshaling", func() {
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
		keyCfg := crypto.Key(key)

		keyCfgStr, err := keyCfg.MarshalText()
		Expect(err).To(Succeed())
		Expect(string(keyCfgStr)).To(Equal(keyStr))
	})

	It("fails on invalid key", func() {
		var k crypto.Key
		err := k.UnmarshalText([]byte(brokenKeyStr))
		Expect(err).To(HaveOccurred())
	})
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
	Expect(kp1.Shared().Bytes()).To(test.BeRandom())
})

var _ = It("can generate valid IPv6 link-local addresses from a public key", func() {
	key, err := crypto.GeneratePrivateKey()
	Expect(err).To(Succeed())

	addr := key.PublicKey().IPv6Address()

	_, ll, err := net.ParseCIDR("fe80::/10")
	Expect(err).To(Succeed())

	ones, bits := addr.Mask.Size()
	Expect(ones).To(Equal(64))
	Expect(bits).To(Equal(128))
	Expect(ll.Contains(addr.IP)).To(BeTrue())
})

var _ = It("can generate valid IPv4 link-local addresses from a public key", func() {
	key, err := crypto.GeneratePrivateKey()
	Expect(err).To(Succeed())

	addr := key.PublicKey().IPv4Address()

	_, ll, err := net.ParseCIDR("169.254.0.0/16")
	Expect(err).To(Succeed())

	ones, bits := addr.Mask.Size()
	Expect(ones).To(Equal(16))
	Expect(bits).To(Equal(32))
	Expect(ll.Contains(addr.IP)).To(BeTrue())
})

var _ = It("can generate a public key pair from a public/private one", func() {
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
