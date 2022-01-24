package crypto

import (
	"crypto/sha512"

	"github.com/Scratch-net/vxeddsa/edwards25519"
	"golang.org/x/crypto/ed25519"
)

// sign signs the message with privateKey and returns a signature as a byte slice.
func (sk Key) Sign(msg []byte, nonce Nonce) Signature {

	// Calculate Ed25519 public key from Curve25519 private key
	var A edwards25519.ExtendedGroupElement
	var pk [32]byte

	edwards25519.GeScalarMultBase(&A, (*[32]byte)(&sk))
	A.ToBytes(&pk)

	// Calculate r
	diversifier := []byte{
		0xFE, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

	var r [64]byte
	hash := sha512.New()
	hash.Write(diversifier)
	hash.Write(sk[:])
	hash.Write(msg)
	hash.Write(nonce[:])
	hash.Sum(r[:0])

	// Calculate R
	var rReduced [32]byte
	edwards25519.ScReduce(&rReduced, &r)
	var R edwards25519.ExtendedGroupElement
	edwards25519.GeScalarMultBase(&R, &rReduced)

	var encr [32]byte
	R.ToBytes(&encr)

	// Calculate S = r + SHA2-512(R || A_ed || msg) * a  (mod L)
	var hramDigest [64]byte
	hash.Reset()
	hash.Write(encr[:])
	hash.Write(pk[:])
	hash.Write(msg)
	hash.Sum(hramDigest[:0])
	var hramDigestReduced [32]byte
	edwards25519.ScReduce(&hramDigestReduced, &hramDigest)

	var s [32]byte
	edwards25519.ScMulAdd(&s, &hramDigestReduced, (*[32]byte)(&sk), &rReduced)

	var sig Signature
	copy(sig[:32], encr[:])
	copy(sig[32:], s[:])
	sig[63] |= pk[31] & 0x80

	return sig
}

// verify checks whether the message has a valid signature.
func (pk Key) Verify(msg []byte, sig Signature) bool {
	pk[31] &= 0x7F

	/* Convert the Curve25519 public key into an Ed25519 public key.  In
	particular, convert Curve25519's "montgomery" x-coordinate into an
	Ed25519 "edwards" y-coordinate:
	ed_y = (mont_x - 1) / (mont_x + 1)
	NOTE: mont_x=-1 is converted to ed_y=0 since fe_invert is mod-exp
	Then move the sign bit into the pubkey from the signature.
	*/

	var edY, one, montX, montXMinusOne, montXPlusOne edwards25519.FieldElement
	edwards25519.FeFromBytes(&montX, (*[32]byte)(&pk))
	edwards25519.FeOne(&one)
	edwards25519.FeSub(&montXMinusOne, &montX, &one)
	edwards25519.FeAdd(&montXPlusOne, &montX, &one)
	edwards25519.FeInvert(&montXPlusOne, &montXPlusOne)
	edwards25519.FeMul(&edY, &montXMinusOne, &montXPlusOne)

	var A_ed [32]byte
	edwards25519.FeToBytes(&A_ed, &edY)

	A_ed[31] |= sig[63] & 0x80
	sig[63] &= 0x7F

	return ed25519.Verify(A_ed[:], msg, sig[:])
}
