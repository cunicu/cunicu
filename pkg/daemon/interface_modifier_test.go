// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package daemon_test

import (
	"math"

	"cunicu.li/cunicu/pkg/daemon"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("interface modifier", func() {
	It("can string", func() {
		for i := range len(daemon.InterfaceModifiersStrings) {
			mod := daemon.InterfaceModifier(1 << i)

			Expect(mod.String()).To(Equal(daemon.InterfaceModifiersStrings[i]))
		}
	})

	It("can strings", func() {
		mod := daemon.InterfaceModifier(math.MaxInt)
		for range len(daemon.InterfaceModifiersStrings) {
			Expect(mod.Strings()).To(Equal(daemon.InterfaceModifiersStrings))
		}
	})

	It("can check if set", func() {
		mod := daemon.InterfaceModifiedPeers

		Expect(mod.Is(daemon.InterfaceModifiedName)).To(BeFalse())
		Expect(mod.Is(daemon.InterfaceModifiedPeers)).To(BeTrue())
	})
})
