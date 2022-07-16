package wg_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"kernel.org/pub/linux/libs/security/libcap/cap"
	"riasc.eu/wice/internal/util"
	"riasc.eu/wice/internal/wg"
)

var _ = It("detects the kernel module", func() {
	if !util.HasCapabilities(cap.NET_ADMIN) {
		Skip("Insufficient privileges")
	}

	Expect(wg.KernelModuleExists()).To(BeTrue())
})
