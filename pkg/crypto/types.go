package crypto

import (
	"crypto/hmac"
	"encoding/base64"

	"github.com/pion/dtls/v2/pkg/crypto/hash"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Keys

const (
	KeyLength       = 32
	SignatureLength = 64
)

type Key [KeyLength]byte
type Signature [SignatureLength]byte

type KeyPair struct {
	Private Key
	Public  Key
}

type PublicKeyPair struct {
	Ours   Key
	Theirs Key
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

// Checks if the key is not zero
func (k Key) IsSet() bool {
	return k != Key{}
}

func (kp PublicKeyPair) ID(key []byte) string {
	ctx := hmac.New(hash.SHA512.CryptoHash().HashFunc().New, key)

	ctx.Write(kp.Ours[:])
	ctx.Write(kp.Theirs[:])

	mac := ctx.Sum(nil)

	return base64.URLEncoding.EncodeToString(mac)
}
