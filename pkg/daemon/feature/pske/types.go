package pske

import kyber "github.com/symbolicsoft/kyber-k2so"

type KyberCipherText [kyber.Kyber1024CTBytes]byte
type KyberPublicKey [kyber.Kyber1024PKBytes]byte
type KyberSecretKey [kyber.Kyber1024SKBytes]byte
