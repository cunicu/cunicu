// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package crypto

// TODO: Remove nolint directive once gci knows the new package
//nolint:gci
import (
	"crypto/ecdh"
	"encoding/base64"
	"errors"
	"math/big"
	"net"

	"github.com/dchest/siphash"
	"golang.org/x/crypto/argon2"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Keys

const (
	KeyLength = 32
)

//nolint:gochecknoglobals
var (
	// A cunīcu specific key for siphash to generate unique IPv6 addresses from the
	// interfaces public key
	addrHashKey = [...]byte{0x67, 0x67, 0x2c, 0x05, 0xd1, 0x3e, 0x11, 0x94, 0xbb, 0x38, 0x91, 0xff, 0x4f, 0x80, 0xb3, 0x97}

	argonSalt = [...]byte{0x77, 0x31, 0x63, 0x33, 0x63, 0x30, 0x6e, 0x6e, 0x33, 0x63, 0x74, 0x73, 0x33, 0x76, 0x65, 0x72, 0x79, 0x62, 0x30, 0x64, 0x79}

	errInvalidKeyLength = errors.New("invalid length")
)

type (
	Nonce []byte
	Key   [KeyLength]byte
)

func GenerateKeyFromPassword(pw string) Key {
	key := argon2.IDKey([]byte(pw), argonSalt[:], 1, 64*1024, 4, KeyLength)

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
		return Key{}, errInvalidKeyLength
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

func (k Key) IPAddress(p net.IPNet) net.IPNet {
	ones, bits := p.Mask.Size()

	hash := siphash.New128(addrHashKey[:])
	if n, err := hash.Write(k[:]); err != nil {
		panic(err)
	} else if n != KeyLength {
		panic("incomplete hash")
	}

	// Append interface identifier from the hash function
	var db []byte
	db = hash.Sum(db)

	n := p.Mask
	b := p.IP
	if c := p.IP.To4(); c != nil {
		b = c
	}

	d := new(big.Int).SetBytes(db[:bits/8])
	m := new(big.Int).SetBytes(n)
	i := new(big.Int).SetBytes(b)

	d.Rsh(d, uint(ones))
	i.And(i, m)
	d.Or(d, i)

	return net.IPNet{
		IP:   d.Bytes(),
		Mask: m.Bytes(),
	}
}

// Checks if the key is not zero
func (k Key) IsSet() bool {
	return k != Key{}
}

// A key which uses GenerateKeyFromPassword() for UnmarshalText()
type KeyPassphrase Key

func (k *KeyPassphrase) UnmarshalText(text []byte) error {
	*k = KeyPassphrase(GenerateKeyFromPassword(string(text)))
	return nil
}

type KeyPair struct {
	Ours   Key `json:"ours"`
	Theirs Key `json:"theirs"`
}

type PublicKeyPair KeyPair

func (kp KeyPair) Shared() Key {
	sk, _ := ecdh.X25519().NewPrivateKey(kp.Ours[:])
	pk, _ := ecdh.X25519().NewPublicKey(kp.Theirs[:])

	shared, err := sk.ECDH(pk)
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
