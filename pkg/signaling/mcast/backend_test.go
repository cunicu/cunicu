// SPDX-FileCopyrightText: 2025 Adam Rizkalla <ajarizzo@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package mcast_test

import (
	"net/url"
	"testing"

	"cunicu.li/cunicu/test"

	_ "cunicu.li/cunicu/pkg/signaling/mcast"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	test.SetupLogging()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Multicast Backend Suite")
}

var _ = Describe("Multicast backend", func() {
	u := url.URL{
		Scheme:   "multicast",
		Host:     "239.0.0.1:9999",
		RawQuery: "interface=lo&loopback=true",
	}

	test.BackendTest(&u, 10)
})
