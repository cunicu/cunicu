package core_test

import (
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"riasc.eu/wice/pkg/core"
)

var _ = Context("interface modifier", func() {
	It("can string", func() {
		for i := 0; i < len(core.InterfaceModifiersStrings); i++ {
			mod := core.InterfaceModifier(1 << i)

			Expect(mod.String()).To(Equal(core.InterfaceModifiersStrings[i]))
		}
	})

	It("can strings", func() {
		mod := core.InterfaceModifier(math.MaxInt)
		for i := 0; i < len(core.InterfaceModifiersStrings); i++ {
			Expect(mod.Strings()).To(Equal(core.InterfaceModifiersStrings))
		}
	})

	It("can check if set", func() {
		mod := core.InterfaceModifiedPeers

		Expect(mod.Is(core.InterfaceModifiedName)).To(BeFalse())
		Expect(mod.Is(core.InterfaceModifiedPeers)).To(BeTrue())
	})
})
