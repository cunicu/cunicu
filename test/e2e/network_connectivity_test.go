// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"context"

	opt "github.com/stv0g/cunicu/test/e2e/nodes/options"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func (n *Network) ConnectivityTests() {
	It("", func() {
		By("Waiting until all peers are connected")

		ctx, cancel := context.WithTimeout(context.Background(), options.timeout)
		defer cancel()

		err := n.AgentNodes.WaitConnectionsReady(ctx)
		Expect(err).To(Succeed(), "Failed to wait for peers to connect: %s", err)

		By("Ping between all peers started")

		err = n.AgentNodes.PingPeers(ctx)
		Expect(err).To(Succeed(), "Failed to ping peers: %s", err)

		By("Ping between all peers succeeded")
	})
}

func (n *Network) ConnectivityTestsWithExtraArgs(extraArgs ...any) {
	BeforeEach(func() {
		n.AgentOptions = append(n.AgentOptions,
			opt.ExtraArgs(extraArgs),
		)
	})

	n.ConnectivityTests()
}

func (n *Network) ConnectivityTestsForAllCandidateTypes() {
	Context("candidate-types", func() {
		Context("any: Allow any candidate type", func() {
			Context("ipv4: Allow IPv4 network only", func() {
				n.ConnectivityTestsWithExtraArgs("--ice-network-type", "udp4")
			})

			Context("ipv6: Allow IPv6 network only", func() {
				n.ConnectivityTestsWithExtraArgs("--ice-network-type", "udp6")
			})
		})

		Context("host: Allow only host candidates", func() {
			Context("ipv4: Allow IPv4 network only", func() {
				n.ConnectivityTestsWithExtraArgs("--ice-candidate-type", "host", "--ice-network-type", "udp4") // , "--port-forwarding=false")
			})

			Context("ipv6: Allow IPv6 network only", func() {
				n.ConnectivityTestsWithExtraArgs("--ice-candidate-type", "host", "--ice-network-type", "udp6")
			})
		})

		Context("srflx: Allow only server reflexive candidates", func() {
			Context("ipv4: Allow IPv4 network only", func() {
				n.ConnectivityTestsWithExtraArgs("--ice-candidate-type", "srflx", "--ice-network-type", "udp4")
			})

			Context("ipv6: Allow IPv6 network only", func() {
				n.ConnectivityTestsWithExtraArgs("--ice-candidate-type", "srflx", "--ice-network-type", "udp6")
			})
		})

		Context("relay: Allow only relay candidates", func() {
			Context("ipv4: Allow IPv4 network only", func() {
				n.ConnectivityTestsWithExtraArgs("--ice-candidate-type", "relay", "--ice-network-type", "udp4")
			})

			// TODO: Check why IPv6 relay is not working
			// Blocked by: https://github.com/pion/ice/pull/462
			Context("ipv6", Pending, func() {
				n.ConnectivityTestsWithExtraArgs("--ice-candidate-type", "relay", "--ice-network-type", "udp6")
			})
		})
	})
}
