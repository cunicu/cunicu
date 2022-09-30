package grpc_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stv0g/cunicu/pkg/signaling/grpc"
)

var _ = Context("relay server", func() {
	DescribeTable("can parse relay urls",
		func(urlStr string, expRelay grpc.RelayInfo) {
			relay, err := grpc.NewRelayInfo(urlStr)
			Expect(err).To(Succeed())
			Expect(relay).To(Equal(expRelay))
		},
		Entry("simple", "stun:stun.cunicu.li", grpc.RelayInfo{
			URL: "stun:stun.cunicu.li:3478",
			TTL: grpc.DefaultRelayTTL,
		}),
		Entry("turn with secret", "turn:turn.cunicu.li?secret=mysecret", grpc.RelayInfo{
			URL:    "turn:turn.cunicu.li:3478?transport=udp",
			Secret: "mysecret",
			TTL:    grpc.DefaultRelayTTL,
		}),
		Entry("turn with user + pass", "turn:user1:pass1@turn.cunicu.li", grpc.RelayInfo{
			URL:      "turn:turn.cunicu.li:3478?transport=udp",
			Username: "user1",
			Password: "pass1",
			TTL:      grpc.DefaultRelayTTL,
		}),
		Entry("turn with user + pass", "turn:turn.cunicu.li?ttl=2h", grpc.RelayInfo{
			URL: "turn:turn.cunicu.li:3478?transport=udp",
			TTL: 2 * time.Hour,
		}),
	)
})
