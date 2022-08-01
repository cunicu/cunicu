package wg_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"riasc.eu/wice/pkg/util"
	"riasc.eu/wice/pkg/wg"
)

var _ = It("detects the kernel module", func() {
	if !util.HasAdminPrivileges() {
		Skip("Insufficient privileges")
	}

	Expect(wg.KernelModuleExists()).To(BeTrue())
})
