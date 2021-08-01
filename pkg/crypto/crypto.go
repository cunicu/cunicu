package crypto

import (
	"math/big"

	"golang.org/x/crypto/curve25519"
)

// endecrypt encrypts a 32-byte slice given the their public & our private private curve25519 keys via simple XOR with the shared secret
func Curve25519Crypt(privKey, pubKey Key, payloadBuf []byte) ([]byte, error) {

	// Perform static-static ECDH
	keyBuf, err := curve25519.X25519(privKey[:], pubKey[:])
	if err != nil {
		return nil, err
	}

	var key, payload, enc big.Int
	key.SetBytes(keyBuf)
	payload.SetBytes(payloadBuf)
	enc.Xor(&key, &payload)

	return enc.Bytes(), nil
}
