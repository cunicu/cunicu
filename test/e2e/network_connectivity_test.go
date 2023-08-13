// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"context"

	opt "cunicu.li/cunicu/test/e2e/nodes/options"

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

func (n *Network) ConnectivityTestsForAllCandidateTypes() {
	ConnectivityTests := func() {
		Context("ipv4: Allow IPv4 network only", func() {
			BeforeEach(func() {
				n.AgentOptions = append(n.AgentOptions, opt.ConfigValue("ice.network_types", "udp4"))
			})

			n.ConnectivityTests()
		})

		Context("ipv6: Allow IPv6 network only", func() {
			BeforeEach(func() {
				n.AgentOptions = append(n.AgentOptions, opt.ConfigValue("ice.network_types", "udp6"))
			})

			n.ConnectivityTests()
		})
	}

	Context("candidate-types", func() {
		Context("any: Allow any candidate type", func() {
			ConnectivityTests()
		})

		Context("host: Allow only host candidates", Label("host"), func() {
			BeforeEach(func() {
				n.AgentOptions = append(n.AgentOptions, opt.ConfigValue("ice.candidate_types", "host"))
			})

			ConnectivityTests()
		})

		Context("srflx: Allow only server reflexive candidates", Label("srflx"), func() {
			BeforeEach(func() {
				n.AgentOptions = append(n.AgentOptions, opt.ConfigValue("ice.candidate_types", "srflx"))
			})

			ConnectivityTests()
		})

		Context("relay: Allow only relay candidates", Label("relay"), func() {
			BeforeEach(func() {
				n.AgentOptions = append(n.AgentOptions, opt.ConfigValue("ice.candidate_types", "relay"))
			})

			Context("ipv4: Allow IPv4 network only", Label("ipv4"), func() {
				BeforeEach(func() {
					n.AgentOptions = append(n.AgentOptions, opt.ConfigValue("ice.network_types", "udp4"))
				})

				n.ConnectivityTests()
			})

			// TODO: Check why IPv6 relay is not working
			// Blocked by: https://github.com/pion/ice/pull/462
			Context("ipv6: Allow IPv6 network only", Label("ipv6"), Pending, func() {
				BeforeEach(func() {
					n.AgentOptions = append(n.AgentOptions, opt.ConfigValue("ice.network_types", "udp6"))
				})

				n.ConnectivityTests()
			})
		})
	})
}
