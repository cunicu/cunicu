// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package ice_test

import (
	"github.com/pion/ice/v3"

	icex "cunicu.li/cunicu/pkg/ice"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("types", func() {
	DescribeTable("can parse candidate types",
		func(str string, want ice.CandidateType) {
			is, err := icex.ParseCandidateType(str)
			Expect(err).To(Succeed())
			Expect(is).To(Equal(want))
		},
		Entry("host", "host", ice.CandidateTypeHost),
		Entry("srflx", "srflx", ice.CandidateTypeServerReflexive),
		Entry("prflx", "prflx", ice.CandidateTypePeerReflexive),
		Entry("relay", "relay", ice.CandidateTypeRelay),
	)

	DescribeTable("can parse network types",
		func(str string, want ice.NetworkType) {
			is, err := icex.ParseNetworkType(str)
			Expect(err).To(Succeed())
			Expect(is).To(Equal(want))
		},
		Entry("udp4", "udp4", ice.NetworkTypeUDP4),
		Entry("udp6", "udp6", ice.NetworkTypeUDP6),
		Entry("tcp4", "tcp4", ice.NetworkTypeTCP4),
		Entry("tcp6", "tcp6", ice.NetworkTypeTCP6),
	)
})
