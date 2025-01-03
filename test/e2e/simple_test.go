// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"

	g "cunicu.li/gont/v2/pkg"
	gopt "cunicu.li/gont/v2/pkg/options"
	copt "cunicu.li/gont/v2/pkg/options/cmd"
	gfopt "cunicu.li/gont/v2/pkg/options/filters"
	"golang.org/x/sys/unix"

	"cunicu.li/cunicu/pkg/config"
	"cunicu.li/cunicu/pkg/wg"
	"cunicu.li/cunicu/test"
	"cunicu.li/cunicu/test/e2e/nodes"
	opt "cunicu.li/cunicu/test/e2e/nodes/options"
	wopt "cunicu.li/cunicu/test/e2e/nodes/options/wg"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

/* Simple local-area switched topology with variable number of agents
 *
 *  - 1x Relay node         [r1] (Coturn STUN/TURN server)
 *  - 1x Signaling node     [s1] (gRPC server)
 *  - 1x Switch             [sw1]
 *  - Yx cunicu Agent nodes [n?]
 *
 *        Relay            Signaling
 *        ┌────┐            ┌────┐
 *        │ r1 │            │ s1 │
 *        └──┬─┘            └─┬──┘
 *           └─────┐   ┌──────┘
 *                ┌┴───┴┐
 *                │ sw1 │ Switch
 *                └┬─┬─┬┘
 *           ┌─────┘ │ └───────┐
 *        ┌──┴─┐  ┌──┴─┐     ┌─┴──┐
 *        │ n1 │  │ n2 │ ... │ nY │
 *        └────┘  └────┘     └────┘
 *               cunicu Agents.
 */
var _ = Context("simple: Simple local-area switched topology with variable number of agents", Serial, func() {
	var (
		err error
		n   Network
		nw  *g.Network

		NumAgents int
	)

	BeforeEach(OncePerOrdered, func() {
		n.Init()

		NumAgents = 3

		n.AgentOptions = append(n.AgentOptions,
			gopt.EmptyDir(wg.ConfigPath),
			gopt.EmptyDir(wg.SocketPath),
		)

		n.WireGuardInterfaceOptions = append(n.WireGuardInterfaceOptions,
			wopt.FullMeshPeers,
		)
	})

	AfterEach(OncePerOrdered, func() {
		n.Close()
	})

	JustBeforeEach(OncePerOrdered, func() {
		By("Initializing core network")

		nw, err = g.NewNetwork(n.Name, n.NetworkOptions...)
		Expect(err).To(Succeed(), "Failed to create network: %s", err)

		sw1, err := nw.AddSwitch("sw1")
		Expect(err).To(Succeed(), "Failed to create switch: %s", err)

		By("Initializing relay node")

		r1, err := nodes.NewCoturnNode(nw, "r1",
			g.NewInterface("eth0", sw1,
				gopt.AddressIP("10.0.0.1/16"),
				gopt.AddressIP("fc::1/64"),
			),
		)
		Expect(err).To(Succeed(), "Failed to start relay: %s", err)

		By("Initializing signaling node")

		s1, err := nodes.NewGrpcSignalingNode(nw, "s1",
			g.NewInterface("eth0", sw1,
				gopt.AddressIP("10.0.0.2/16"),
				gopt.AddressIP("fc::2/64"),
			),
		)
		Expect(err).To(Succeed(), "Failed to create signaling node: %s", err)

		By("Initializing agent nodes")

		n.AgentNodes, err = test.ParallelNew(NumAgents, func(i int) (*nodes.Agent, error) {
			return nodes.NewAgent(nw, fmt.Sprintf("n%d", i),
				gopt.Customize[g.Option](n.AgentOptions,
					g.NewInterface("eth0", sw1,
						gopt.AddressIP("10.0.1.%d/16", i),
						gopt.AddressIP("fc::1:%d/64", i),
					),
					wopt.Interface("wg0",
						gopt.Customize[g.Option](n.WireGuardInterfaceOptions,
							wopt.AddressIP("172.16.0.%d/16", i),
						)...,
					),
				)...,
			)
		})
		Expect(err).To(Succeed(), "Failed to create agent nodes: %s", err)

		By("Starting network")

		n.Network = nw
		n.RelayNodes = nodes.RelayList{r1}
		n.SignalingNodes = nodes.SignalingList{s1}

		n.Start()
	})

	Context("kernel: Use kernel WireGuard interface", Label("kernel"), func() {
		n.ConnectivityTestsForAllCandidateTypes()
	})

	Context("userspace: Use wireguard-go userspace interfaces", Label("userspace"), Pending, func() {
		BeforeEach(func() {
			n.WireGuardInterfaceOptions = append(n.WireGuardInterfaceOptions,
				wopt.WriteConfigFile(true),
				wopt.SetupKernelInterface(false),
			)

			n.AgentOptions = append(n.AgentOptions,
				opt.ConfigValue("interfaces.wg0", config.InterfaceSettings{
					UserSpace: true,
				}),
			)
		})

		n.ConnectivityTestsForAllCandidateTypes()
	})

	Context("no-nat: Disable NAT for kernel device", Pending, func() {
	})

	Context("filtered: Block WireGuard UDP traffic", func() {
		Context("p2p: Between agents only", func() {
			BeforeEach(func() {
				// We are dropped packets between the cunīcu nodes to force ICE using the relay
				n.AgentOptions = append(n.AgentOptions,
					gopt.Filter(g.FilterInput,
						gfopt.InputInterfaceName("eth0"),
						gfopt.SourceIP("10.0.1.0/24"),
						gfopt.Drop,
					),
					gopt.Filter(g.FilterInput,
						gfopt.InputInterfaceName("eth0"),
						gfopt.SourceIP("fc::1:0/112"),
						gfopt.Drop,
					),
				)
			})

			n.ConnectivityTests()
		})

		Context("all-udp: All UDP entirely", func() {
			BeforeEach(func() {
				n.AgentOptions = append(n.AgentOptions,
					gopt.Filter(g.FilterInput,
						gfopt.InputInterfaceName("eth0"),
						gfopt.TransportProtocol(unix.IPPROTO_UDP),
						gfopt.Drop,
					),
				)
			})

			n.ConnectivityTests()
		})
	})

	Context("pdisc: Peer Discovery", Pending, Ordered, func() {
		BeforeEach(OncePerOrdered, func() {
			n.AgentOptions = append(n.AgentOptions,
				opt.ConfigValue("community", "hallo"),
			)

			n.WireGuardInterfaceOptions = append(n.WireGuardInterfaceOptions,
				wopt.NoPeers,
			)

			NumAgents = 3
		})

		n.ConnectivityTests()

		It("", func() {
			By("Check existing peers")

			err := n.AgentNodes.ForEachAgent(func(a *nodes.Agent) error {
				_, err := a.Run("wg", copt.Combined(GinkgoWriter))
				Expect(err).To(Succeed())

				return nil
			})
			Expect(err).To(Succeed())
		})
	})
})
