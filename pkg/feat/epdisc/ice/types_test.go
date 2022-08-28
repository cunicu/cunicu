package ice_test

import (
	"github.com/pion/ice/v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	icex "riasc.eu/wice/pkg/feat/epdisc/ice"
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
			var ctp = icex.CandidateType{ct}
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
			var ntp = icex.NetworkType{ct}
			m, err := ntp.MarshalText()
			Expect(err).To(Succeed())
			Expect(string(m)).To(Equal(st))
		}, t)
	})

	Context("ICE URL", func() {
		t := []TableEntry{
			Entry(nil, "stun:example.com", "stun:example.com:3478", ice.URL{Scheme: ice.SchemeTypeSTUN, Host: "example.com", Port: 3478, Username: "", Password: "", Proto: ice.ProtoTypeUDP}),
			Entry(nil, "stuns:example.com", "stuns:example.com:5349", ice.URL{Scheme: ice.SchemeTypeSTUNS, Host: "example.com", Port: 5349, Username: "", Password: "", Proto: ice.ProtoTypeTCP}),
			Entry(nil, "stun:example.com:1234", "stun:example.com:1234", ice.URL{Scheme: ice.SchemeTypeSTUN, Host: "example.com", Port: 1234, Username: "", Password: "", Proto: ice.ProtoTypeUDP}),
			Entry(nil, "stuns:example.com:1234", "stuns:example.com:1234", ice.URL{Scheme: ice.SchemeTypeSTUNS, Host: "example.com", Port: 1234, Username: "", Password: "", Proto: ice.ProtoTypeTCP}),
			Entry(nil, "turn:example.com?transport=tcp", "turn:example.com:3478?transport=tcp", ice.URL{Scheme: ice.SchemeTypeTURN, Host: "example.com", Port: 3478, Username: "", Password: "", Proto: ice.ProtoTypeTCP}),
			Entry(nil, "turns:example.com", "turns:example.com:5349?transport=tcp", ice.URL{Scheme: ice.SchemeTypeTURNS, Host: "example.com", Port: 5349, Username: "", Password: "", Proto: ice.ProtoTypeTCP}),
		}

		DescribeTable("Unmarshal", func(u, _ string, e ice.URL) {
			var up icex.URL
			Expect(up.UnmarshalText([]byte(u))).To(Succeed())
			Expect(up.URL).To(Equal(e))
		}, t)

		DescribeTable("Marshal", func(_, u string, e ice.URL) {
			var up = icex.URL{e}
			m, err := up.MarshalText()
			Expect(err).To(Succeed())
			Expect(string(m)).To(Equal(u))
		}, t)
	})
})
