package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"net"

	"github.com/aead/siphash"
	"github.com/pion/dtls/v2/pkg/crypto/hash"
	"golang.org/x/crypto/pbkdf2"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Keys

const (
	KeyLength       = 32
	SignatureLength = 64

	pbkdf2Iterations = 4096
)

var (
	// A WICE specific key for siphash to generate unique IPv6 addresses from the
	// interfaces public key
	addrHashKey = [...]byte{0x67, 0x67, 0x2c, 0x05, 0xd1, 0x3e, 0x11, 0x94, 0xbb, 0x38, 0x91, 0xff, 0x4f, 0x80, 0xb3, 0x97}

	pbkdf2Salt = [...]byte{0x77, 0x31, 0x63, 0x33, 0x63, 0x30, 0x6e, 0x6e, 0x33, 0x63, 0x74, 0x73, 0x33, 0x76, 0x65, 0x72, 0x79, 0x62, 0x30, 0x64, 0x79}
)

type Nonce []byte
type Key [KeyLength]byte
type Signature [SignatureLength]byte

type KeyPair struct {
	Ours   Key `json:"ours"`
	Theirs Key `json:"theirs"`
}

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

	var key Key
	copy(key[:], k[:KeyLength])

	return key, nil
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

	hash, _ := siphash.New64(addrHashKey[:])
	hash.Write(k[:])

	// Append interface identifier from the hash function
	ip = hash.Sum(ip)

	return &net.IPNet{
		IP:   ip,
		Mask: net.CIDRMask(64, 128),
	}
}

// Checks if the key is not zero
func (k Key) IsSet() bool {
	return k != Key{}
}

func (k Signature) String() string {
	return base64.StdEncoding.EncodeToString(k[:])
}

func (kp KeyPair) ID(key []byte) string {
	ctx := hmac.New(hash.SHA512.CryptoHash().HashFunc().New, key)

	ctx.Write(kp.Ours[:])
	ctx.Write(kp.Theirs[:])

	mac := ctx.Sum(nil)

	return base64.URLEncoding.EncodeToString(mac)
}

func (kp KeyPair) Shared() Key {
	shared := Key{}

	for i := range kp.Ours {
		shared[i] = kp.Ours[i] ^ kp.Theirs[i]
	}

	return shared
}

func GetNonce(len int) (Nonce, error) {
	var nonce = make(Nonce, len)

	_, err := rand.Reader.Read(nonce)
	if err != nil {
		return nonce, err
	}

	return nonce, nil
}
