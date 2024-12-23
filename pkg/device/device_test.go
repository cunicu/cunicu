// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package device_test

import (
	"testing"

	osx "cunicu.li/cunicu/pkg/os"
	"cunicu.li/cunicu/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	test.SetupLogging()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Device Suite")
}

var _ = BeforeSuite(func() {
	if !osx.HasAdminPrivileges() {
		Skip("Insufficient privileges")
	}
})
