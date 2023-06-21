// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package inprocess_test

import (
	"net/url"
	"testing"

	"github.com/stv0g/cunicu/test"

	_ "github.com/stv0g/cunicu/pkg/signaling/inprocess"

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
