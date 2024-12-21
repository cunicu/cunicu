// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package crypto implements basic crypto primitives used in the project
package crypto

import (
	"crypto/rand"
	"math/big"
)

func GetNonce(length int) (Nonce, error) {
	nonce := make(Nonce, length)

	_, err := rand.Read(nonce)
	if err != nil {
		return nonce, err
	}

	return nonce, nil
}

// GetRandomString generates a random string for cryptographic usage.
func GetRandomString(n int, runes string) (string, error) {
	letters := []rune(runes)
	b := make([]rune, n)
	for i := range b {
		v, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		b[i] = letters[v.Int64()]
	}
	return string(b), nil
}
