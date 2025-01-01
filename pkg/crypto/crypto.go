// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package crypto implements basic crypto primitives used in the project
package crypto

import (
	"crypto/rand"
	"math/big"
)

// Intn is a shortcut for generating a random integer between 0 and
// max using crypto/rand.
func Intn(maxi int64) int64 {
	nBig, err := rand.Int(rand.Reader, big.NewInt(maxi))
	if err != nil {
		panic(err)
	}

	return nBig.Int64()
}

// Float64 is a shortcut for generating a random float between 0 and
// 1 using crypto/rand.
func Float64() float64 {
	return float64(Intn(1<<53)) / (1 << 53)
}

func GetNonce(length int) (Nonce, error) {
	nonce := make(Nonce, length)

	_, err := rand.Read(nonce)
	if err != nil {
		return nonce, err
	}

	return nonce, nil
}

// Returns a random value from the following interval:
//
//	[currentInterval - randomizationFactor * currentInterval, currentInterval + randomizationFactor * currentInterval].
func GetRandomValueFromInterval[V ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | float32 | float64](randomizationFactor float64, currentInterval V) V {
	if randomizationFactor == 0 {
		return currentInterval // Make sure no randomness is used when randomizationFactor is 0.
	}

	delta := randomizationFactor * float64(currentInterval)

	minInterval := float64(currentInterval) - delta
	maxInterval := float64(currentInterval) + delta

	// Get a random value from the range [minInterval, maxInterval].
	// The formula used below has a +1 because if the minInterval is 1 and the maxInterval is 3 then
	// we want a 33% chance for selecting either 1, 2 or 3.
	return V(minInterval + (Float64() * (maxInterval - minInterval + 1)))
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
