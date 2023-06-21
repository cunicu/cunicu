// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package test implements universal helpers for unit and integration tests
package test

import (
	"bytes"
	"crypto/rand"
	"math"
	"math/big"
	"os"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/stv0g/cunicu/pkg/crypto"
	epdiscproto "github.com/stv0g/cunicu/pkg/proto/feature/epdisc"
	"github.com/stv0g/cunicu/pkg/signaling"
)

func GenerateKeyPairs() (*crypto.KeyPair, *crypto.KeyPair, error) {
	ourKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		return nil, nil, err
	}

	theirKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		return nil, nil, err
	}

	return &crypto.KeyPair{
			Ours:   ourKey,
			Theirs: theirKey.PublicKey(),
		}, &crypto.KeyPair{
			Ours:   theirKey,
			Theirs: ourKey.PublicKey(),
		}, nil
}

func GenerateSignalingMessage() *signaling.Message {
	r, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		panic(err)
	}

	return &signaling.Message{
		Candidate: &epdiscproto.Candidate{
			Port: int32(r.Int64()),
		},
	}
}

func Entropy(data []byte) float64 {
	if len(data) == 0 {
		return 0
	}

	length := float64(len(data))
	entropy := 0.0

	for i := 0; i < 256; i++ {
		if p := float64(bytes.Count(data, []byte{byte(i)})) / length; p > 0 {
			entropy += -p * math.Log2(p)
		}
	}

	return entropy
}

func IsCI() bool {
	return os.Getenv("CI") == "true"
}

func ParallelNew[T any](cnt int, ctor func(i int) (T, error)) ([]T, error) {
	ts := []T{}
	mu := sync.Mutex{}

	eg := errgroup.Group{}
	for i := 1; i <= cnt; i++ {
		i := i

		eg.Go(func() error {
			t, err := ctor(i)
			if err != nil {
				return err
			}

			mu.Lock()
			ts = append(ts, t)
			mu.Unlock()

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return ts, nil
}
