// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package inprocess_test

import (
	"net/url"
	"testing"

	"cunicu.li/cunicu/test"

	_ "cunicu.li/cunicu/pkg/signaling/inprocess"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	test.SetupLogging()
	RegisterFailHandler(Fail)
	RunSpecs(t, "In-Process Backend Suite")
}

var _ = Describe("inprocess backend", func() {
	u := url.URL{
		Scheme: "inprocess",
	}

	test.BackendTest(&u, 10)
})
