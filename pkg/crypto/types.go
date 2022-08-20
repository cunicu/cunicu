package crypto

import (
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"net"

	"github.com/aead/siphash"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/pbkdf2"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Keys

const (
	KeyLength = 32

	pbkdf2Iterations = 4096
)

var (
	// A ɯice specific key for siphash to generate unique IPv6 addresses from the
	// interfaces public key
	addrHashKey = [...]byte{0x67, 0x67, 0x2c, 0x05, 0xd1, 0x3e, 0x11, 0x94, 0xbb, 0x38, 0x91, 0xff, 0x4f, 0x80, 0xb3, 0x97}

	pbkdf2Salt = [...]byte{0x77, 0x31, 0x63, 0x33, 0x63, 0x30, 0x6e, 0x6e, 0x33, 0x63, 0x74, 0x73, 0x33, 0x76, 0x65, 0x72, 0x79, 0x62, 0x30, 0x64, 0x79}
)

type Nonce []byte
type Key [KeyLength]byte

func GenerateKeyFromPassword(pw string) Key {
	key := pbkdf2.Key([]byte(pw), pbkdf2Salt[:], pbkdf2Iterations, KeyLength, sha512.New)

	// Modify random bytes using algorithm described at:
	// https://cr.yp.to/ecdh.html.
	key[0] &= 248
	key[31] &= 127
	key[31] |= 64

	return *(*Key)(key)
}

func GenerateKey() (Key, error) {
	key, err := wgtypes.GenerateKey()
	if err != nil {
		return Key{}, err
	}

	return Key(key), nil
}

func GeneratePrivateKey() (Key, error) {
	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return Key{}, err
	}

	return Key(key), nil
}

func ParseKey(str string) (Key, error) {
	k, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return Key{}, err
	}

	return ParseKeyBytes(k)
}

func ParseKeyBytes(buf []byte) (Key, error) {
	if len(buf) != KeyLength {
		return Key{}, errors.New("invalid length")
	}

	return *(*Key)(buf), nil
}

func (k Key) String() string {
	return base64.StdEncoding.EncodeToString(k[:])
}

func (k Key) MarshalText() ([]byte, error) {
	return []byte(k.String()), nil
}

func (k *Key) UnmarshalText(text []byte) error {
	var err error
	*k, err = ParseKey(string(text))
	return err
}

func (k Key) PublicKey() Key {
	key := wgtypes.Key(k)

	return Key(key.PublicKey())
}

func (k Key) Bytes() []byte {
	return k[:]
}

// IPv6Address derives an IPv6 link local address from they key
func (k Key) IPv6Address() *net.IPNet {
	ip := net.IP{0xfe, 0x80, 0, 0, 0, 0, 0, 0}

	hash, err := siphash.New64(addrHashKey[:])
	if err != nil {
		panic(err)
	}

	if n, err := hash.Write(k[:]); err != nil {
		panic(err)
	} else if n != KeyLength {
		panic("incomplete hash")
	}

	// Append interface identifier from the hash function
	ip = hash.Sum(ip)

	if len(ip) != net.IPv6len {
		panic("invalid IP length")
	}

	return &net.IPNet{
		IP:   ip,
		Mask: net.CIDRMask(64, 128),
	}
}

// IPv4Address derives an IPv4 link local address from they key
func (k Key) IPv4Address() *net.IPNet {
	hash, err := siphash.New64(addrHashKey[:])
	if err != nil {
		panic(err)
	}

	if _, err := hash.Write(k[:]); err != nil {
		panic(err)
	}

	sum := hash.Sum64()
	sum2 := uint64(0)

	for i := 0; i < 4; i++ {
		sum2 ^= sum & 0xffff
		sum >>= 16
	}

	c := byte(sum2 >> 8)
	d := byte(sum2 & 0xff)

	// Exclude reserved addresses
	// See: https://datatracker.ietf.org/doc/html/rfc3927#section-2.1
	if c == 0 {
		c++
	} else if c == 0xff {
		c--
	}

	return &net.IPNet{
		IP:   net.IPv4(169, 254, c, d),
		Mask: net.CIDRMask(16, 32),
	}
}

// Checks if the key is not zero
func (k Key) IsSet() bool {
	return k != Key{}
}

type KeyPair struct {
	Ours   Key `json:"ours"`
	Theirs Key `json:"theirs"`
}

type PublicKeyPair KeyPair

func (kp KeyPair) Shared() Key {
	shared, err := curve25519.X25519(kp.Ours[:], kp.Theirs[:])
	if err != nil {
		panic(err)
	}

	return *(*Key)(shared)
}

func (kp KeyPair) Public() PublicKeyPair {
	return PublicKeyPair{
		Ours:   kp.Ours.PublicKey(),
		Theirs: kp.Theirs,
	}
}
