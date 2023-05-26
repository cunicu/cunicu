// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package crypto implements basic crypto primitives used in the project
package crypto

import "crypto/rand"

func GetNonce(length int) (Nonce, error) {
	nonce := make(Nonce, length)

	_, err := rand.Read(nonce)
	if err != nil {
		return nonce, err
	}

	return nonce, nil
}
