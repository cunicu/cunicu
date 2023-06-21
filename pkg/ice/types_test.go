// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package ice_test

import (
	"github.com/pion/ice/v2"
	"github.com/pion/stun"

	icex "github.com/stv0g/cunicu/pkg/ice"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Marshaling of ICE types", func() {
	Context("Candidate type", func() {
		t := []TableEntry{
			Entry("Host", ice.CandidateTypeHost, "host"),
			Entry("ServerReflexive", ice.CandidateTypeServerReflexive, "srflx"),
			Entry("PeerReflexive", ice.CandidateTypePeerReflexive, "prflx"),
			Entry("Relay", ice.CandidateTypeRelay, "relay"),
		}

		DescribeTable("Unmarshal", func(ct ice.CandidateType, st string) {
			var ctp icex.CandidateType
			Expect(ctp.UnmarshalText([]byte(st))).To(Succeed())
			Expect(ctp.CandidateType).To(Equal(ct))
		}, t)

		DescribeTable("Marshal", func(ct ice.CandidateType, st string) {
			ctp := icex.CandidateType{ct}
			m, err := ctp.MarshalText()
			Expect(err).To(Succeed())
			Expect(string(m)).To(Equal(st))
		}, t)
	})

	Context("Network type", func() {
		t := []TableEntry{
			Entry("TCP4", ice.NetworkTypeTCP4, "tcp4"),
			Entry("TCP6", ice.NetworkTypeTCP6, "tcp6"),
			Entry("UDP4", ice.NetworkTypeUDP4, "udp4"),
			Entry("UDP6", ice.NetworkTypeUDP6, "udp6"),
		}

		DescribeTable("Unmarshal", func(ct ice.NetworkType, st string) {
			var ntp icex.NetworkType
			Expect(ntp.UnmarshalText([]byte(st))).To(Succeed())
			Expect(ntp.NetworkType).To(Equal(ct))
		}, t)

		DescribeTable("Marshal", func(ct ice.NetworkType, st string) {
			ntp := icex.NetworkType{ct}
			m, err := ntp.MarshalText()
			Expect(err).To(Succeed())
			Expect(string(m)).To(Equal(st))
		}, t)
	})

	Context("URL", func() {
		t := []TableEntry{
			Entry("stun", "stun:cunicu.0l.de:1234", stun.URI{
				Scheme: stun.SchemeTypeSTUN,
				Host:   "cunicu.0l.de",
				Port:   1234,
				Proto:  stun.ProtoTypeUDP,
			}),
			Entry("stuns", "stuns:cunicu.0l.de:1234", stun.URI{
				Scheme: stun.SchemeTypeSTUNS,
				Host:   "cunicu.0l.de",
				Port:   1234,
				Proto:  stun.ProtoTypeTCP,
			}),
			Entry("turn-udp", "turn:cunicu.0l.de:1234?transport=udp", stun.URI{
				Scheme: stun.SchemeTypeTURN,
				Host:   "cunicu.0l.de",
				Port:   1234,
				Proto:  stun.ProtoTypeUDP,
			}),
			Entry("turn-tcp", "turn:cunicu.0l.de:1234?transport=tcp", stun.URI{
				Scheme: stun.SchemeTypeTURN,
				Host:   "cunicu.0l.de",
				Port:   1234,
				Proto:  stun.ProtoTypeTCP,
			}),
			Entry("turns", "turns:cunicu.0l.de:1234?transport=tcp", stun.URI{
				Scheme: stun.SchemeTypeTURNS,
				Host:   "cunicu.0l.de",
				Port:   1234,
				Proto:  stun.ProtoTypeTCP,
			}),
		}

		DescribeTable("Unmarshal", func(urlStr string, url stun.URI) {
			var u icex.URL
			Expect(u.UnmarshalText([]byte(urlStr))).To(Succeed())
			Expect(u.URL).To(Equal(url))
		}, t)

		DescribeTable("Marshal", func(urlStr string, url stun.URI) {
			u := icex.URL{url}
			m, err := u.MarshalText()
			Expect(err).To(Succeed())
			Expect(string(m)).To(BeEquivalentTo(urlStr))
		}, t)
	})
})
