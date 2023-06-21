// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package wg_test

import (
	osx "github.com/stv0g/cunicu/pkg/os"
	"github.com/stv0g/cunicu/pkg/wg"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = It("detects the kernel module", func() {
	if !osx.HasAdminPrivileges() {
		Skip("Insufficient privileges")
	}

	Expect(wg.KernelModuleExists()).To(BeTrue())
})
