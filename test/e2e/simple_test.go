// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"

	g "github.com/stv0g/gont/v2/pkg"
	gopt "github.com/stv0g/gont/v2/pkg/options"
	copt "github.com/stv0g/gont/v2/pkg/options/cmd"
	gfopt "github.com/stv0g/gont/v2/pkg/options/filters"
	"golang.org/x/sys/unix"

	"github.com/stv0g/cunicu/pkg/wg"
	"github.com/stv0g/cunicu/test"
	"github.com/stv0g/cunicu/test/e2e/nodes"
	opt "github.com/stv0g/cunicu/test/e2e/nodes/options"
	wopt "github.com/stv0g/cunicu/test/e2e/nodes/options/wg"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

/* Simple local-area switched topology with variable number of agents
 *
 *  - 1x Relay node        [r1] (Coturn STUN/TURN server)
 *  - 1x Signaling node    [s1] (GRPC server)
 *  - 1x Switch            [sw1]
 *  - Yx  cunicu Agent nodes [n?]
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
 *               cunicu Agents
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

	ConnectivityTestsWithExtraArgs := func(extraArgs ...any) {
		BeforeEach(func() {
			n.AgentOptions = append(n.AgentOptions,
				opt.ExtraArgs(extraArgs),
			)
		})

		n.ConnectivityTests()
	}

	ConnectivityTestsForAllCandidateTypes := func() {
		Context("candidate-types", func() {
			Context("any: Allow any candidate type", func() {
				Context("ipv4: Allow IPv4 network only", func() {
					ConnectivityTestsWithExtraArgs("--ice-network-type", "udp4")
				})

				Context("ipv6: Allow IPv6 network only", func() {
					ConnectivityTestsWithExtraArgs("--ice-network-type", "udp6")
				})
			})

			Context("host: Allow only host candidates", func() {
				Context("ipv4: Allow IPv4 network only", func() {
					ConnectivityTestsWithExtraArgs("--ice-candidate-type", "host", "--ice-network-type", "udp4") // , "--port-forwarding=false")
				})

				Context("ipv6: Allow IPv6 network only", func() {
					ConnectivityTestsWithExtraArgs("--ice-candidate-type", "host", "--ice-network-type", "udp6")
				})
			})

			Context("srflx: Allow only server reflexive candidates", func() {
				Context("ipv4: Allow IPv4 network only", func() {
					ConnectivityTestsWithExtraArgs("--ice-candidate-type", "srflx", "--ice-network-type", "udp4")
				})

				Context("ipv6: Allow IPv6 network only", func() {
					ConnectivityTestsWithExtraArgs("--ice-candidate-type", "srflx", "--ice-network-type", "udp6")
				})
			})

			Context("relay: Allow only relay candidates", func() {
				Context("ipv4: Allow IPv4 network only", func() {
					ConnectivityTestsWithExtraArgs("--ice-candidate-type", "relay", "--ice-network-type", "udp4")
				})

				// TODO: Check why IPv6 relay is not working
				// Blocked by: https://github.com/pion/ice/pull/462
				Context("ipv6", Pending, func() {
					ConnectivityTestsWithExtraArgs("--ice-candidate-type", "relay", "--ice-network-type", "udp6")
				})
			})
		})
	}

	Context("kernel: Use kernel WireGuard interface", func() {
		ConnectivityTestsForAllCandidateTypes()
	})

	PContext("userspace: Use wireguard-go userspace interfaces", func() {
		BeforeEach(func() {
			n.WireGuardInterfaceOptions = append(n.WireGuardInterfaceOptions,
				wopt.WriteConfigFile(true),
				wopt.SetupKernelInterface(false),
			)

			n.AgentOptions = append(n.AgentOptions,
				opt.ExtraArgs{"--wg-userspace", "wg0"},
			)
		})

		ConnectivityTestsForAllCandidateTypes()
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
				opt.ExtraArgs{"--community", "hallo"},
			)

			n.WireGuardInterfaceOptions = append(n.WireGuardInterfaceOptions,
				wopt.NoPeers,
			)

			NumAgents = 3
		})

		n.ConnectivityTests()

		It("", func() {
			By("Check existing peers 2")

			err := n.AgentNodes.ForEachAgent(func(a *nodes.Agent) error {
				_, err := a.Run("wg", copt.Combined(GinkgoWriter))
				Expect(err).To(Succeed())

				return nil
			})
			Expect(err).To(Succeed())
		})
	})
})
