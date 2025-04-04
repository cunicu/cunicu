// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package wg_test

import (
	osx "cunicu.li/cunicu/pkg/os"
	"cunicu.li/cunicu/pkg/wg"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = It("detects the kernel module", func() {
	if !osx.HasAdminPrivileges() {
		Skip("Insufficient privileges")
	}

	Expect(wg.KernelModuleExists()).To(BeTrue())
})
