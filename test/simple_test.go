//go:build linux

package test_test

import (
	"golang.org/x/sys/unix"
	"riasc.eu/wice/pkg/wg"
	"riasc.eu/wice/test/nodes"
	opt "riasc.eu/wice/test/nodes/options"
	wopt "riasc.eu/wice/test/nodes/options/wg"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	g "github.com/stv0g/gont/pkg"
	gopt "github.com/stv0g/gont/pkg/options"
	gfopt "github.com/stv0g/gont/pkg/options/filters"
)

/* Simple local-area switched topology
 *
 * Hosts:
 *  - 1x Relay node (Coturn STUN/TURN server)
 *  - 1x Signaling node (GRPC server)
 *  - 1x Switch
 *  - 2x wice Agent nodes
 *
 *        ┌────┐   ┌────┐
 *  Relay │ r1 │   │ s1 │ Signaling
 *        └──┬─┘   └─┬──┘
 *           └─┐   ┌─┘
 *            ┌┴───┴┐
 *            │ sw1 │ Switch
 *            └┬───┬┘
 *           ┌─┘   └─┐
 *        ┌──┴─┐   ┌─┴──┐
 *        │ n1 │   │ n2 │ wice Agents
 *        └────┘   └────┘
 */
var _ = Context("simple", Serial, func() {
	var n Network

	BeforeEach(func() {
		n.Init()

		n.AgentOptions = append(n.AgentOptions,
			gopt.EmptyDir(wg.ConfigPath),
			gopt.EmptyDir(wg.SocketPath),
		)
	})

	AfterEach(func() {
		n.Close()
	})

	JustBeforeEach(func() {
		By("Initializing core network")

		nw, err := g.NewNetwork(n.Name, n.NetworkOptions...)
		Expect(err).To(Succeed(), "Failed to create network: %s", err)

		sw1, err := nw.AddSwitch("sw1")
		Expect(err).To(Succeed(), "Failed to create switch: %s", err)

		By("Initializing relay node")

		r1, err := nodes.NewCoturnNode(nw, "r1",
			gopt.Interface("eth0", sw1,
				gopt.AddressIP("10.0.0.1/16"),
				gopt.AddressIP("fc::1/64"),
			),
		)
		Expect(err).To(Succeed(), "Failed to start relay: %s", err)

		By("Initializing signaling node")

		s1, err := nodes.NewGrpcSignalingNode(nw, "s1",
			gopt.Interface("eth0", sw1,
				gopt.AddressIP("10.0.0.2/16"),
				gopt.AddressIP("fc::2/64"),
			),
		)
		Expect(err).To(Succeed(), "Failed to create signaling node: %s", err)

		By("Initializing agent nodes")

		wopts1 := gopt.Customize(n.WireGuardInterfaceOptions,
			wopt.AddressIP("172.16.0.1/16"),
			wopt.PeerFromNames("n2", "wg0",
				wopt.AllowedIPStr("172.16.0.2/32"),
			),
		)

		aopts1 := gopt.Customize(n.AgentOptions,
			gopt.Interface("eth0", sw1,
				gopt.AddressIP("10.0.1.1/16"),
				gopt.AddressIP("fc::1:1/64"),
			),
			wopt.Interface("wg0", wopts1...),
		)

		n1, err := nodes.NewAgent(nw, "n1", aopts1...)
		Expect(err).To(Succeed(), "Failed to create agent node: %s", err)

		wopts2 := gopt.Customize(n.WireGuardInterfaceOptions,
			wopt.AddressIP("172.16.0.2/16"),
			wopt.PeerFromNames("n1", "wg0",
				wopt.AllowedIPStr("172.16.0.1/32"),
			),
		)

		aopts2 := gopt.Customize(n.AgentOptions,
			gopt.Interface("eth0", sw1,
				gopt.AddressIP("10.0.1.2/16"),
				gopt.AddressIP("fc::1:2/64"),
			),
			wopt.Interface("wg0", wopts2...),
		)

		n2, err := nodes.NewAgent(nw, "n2", aopts2...)
		Expect(err).To(Succeed(), "Failed to created nodes: %s", err)

		By("Starting network")

		n.Network = nw
		n.RelayNodes = nodes.RelayList{r1}
		n.SignalingNodes = nodes.SignalingList{s1}
		n.AgentNodes = nodes.AgentList{n1, n2}

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
			Context("any", func() {
				Context("ipv4", func() {
					ConnectivityTestsWithExtraArgs("--ice-network-type", "udp4")
				})

				Context("ipv6", func() {
					ConnectivityTestsWithExtraArgs("--ice-network-type", "udp6")
				})
			})

			Context("host", func() {
				Context("ipv4", func() {
					ConnectivityTestsWithExtraArgs("--ice-candidate-type", "host", "--ice-network-type", "udp4")
				})

				Context("ipv6", func() {
					ConnectivityTestsWithExtraArgs("--ice-candidate-type", "host", "--ice-network-type", "udp6")
				})
			})

			Context("srflx", func() {
				Context("ipv4", func() {
					ConnectivityTestsWithExtraArgs("--ice-candidate-type", "srflx", "--ice-network-type", "udp4")
				})

				Context("ipv6", func() {
					ConnectivityTestsWithExtraArgs("--ice-candidate-type", "srflx", "--ice-network-type", "udp6")
				})
			})

			Context("relay", func() {
				Context("ipv4", func() {
					ConnectivityTestsWithExtraArgs("--ice-candidate-type", "relay", "--ice-network-type", "udp4")
				})

				// TODO: Check why IPv6 relay is not working
				Context("ipv6", Pending, func() {
					ConnectivityTestsWithExtraArgs("--ice-candidate-type", "relay", "--ice-network-type", "udp6")
				})
			})
		})
	}

	Context("kernel", func() {
		ConnectivityTestsForAllCandidateTypes()
	})

	Context("userspace", func() {
		Context("ipv4", func() {
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
	})

	Context("filtered", func() {
		Context("p2p", func() {
			BeforeEach(func() {
				// We are dropped packets between the ɯice nodes to force ICE using the relay
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

		Context("all-udp", func() {
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
})
