package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"encoding/base64"
	"net"

	"github.com/aead/siphash"
	"github.com/pion/dtls/v2/pkg/crypto/hash"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Keys

const (
	NonceLength     = 64
	KeyLength       = 32
	SignatureLength = 64
)

var (
	// A WICE specific key for siphash to generate unique IPv6 addresses from the
	// interfaces public key
	addrHashKey = []byte{0x67, 0x67, 0x2c, 0x05, 0xd1, 0x3e, 0x11, 0x94, 0xbb, 0x38, 0x91, 0xff, 0x4f, 0x80, 0xb3, 0x97}
)

type Nonce [NonceLength]byte
type Key [KeyLength]byte
type Signature [SignatureLength]byte

type KeyPair struct {
	Private Key
	Public  Key
}

type PublicKeyPair struct {
	Ours   Key `json:"ours"`
	Theirs Key `json:"theirs"`
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

	hash, _ := siphash.New64(addrHashKey)
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

func (kp PublicKeyPair) ID(key []byte) string {
	ctx := hmac.New(hash.SHA512.CryptoHash().HashFunc().New, key)

	ctx.Write(kp.Ours[:])
	ctx.Write(kp.Theirs[:])

	mac := ctx.Sum(nil)

	return base64.URLEncoding.EncodeToString(mac)
}

func (kp PublicKeyPair) Shared() Key {
	shared := Key{}

	for i := range kp.Ours {
		shared[i] = kp.Ours[i] ^ kp.Theirs[i]
	}

	return shared
}

func GetNonce() (Nonce, error) {
	var nonce Nonce

	_, err := rand.Reader.Read(nonce[:])
	if err != nil {
		return nonce, err
	}

	return nonce, nil
}
