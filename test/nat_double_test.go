//go:build linux

package test_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	g "github.com/stv0g/gont/pkg"
	gopt "github.com/stv0g/gont/pkg/options"
	"riasc.eu/wice/test/nodes"
	wopt "riasc.eu/wice/test/nodes/options/wg"
)

/* Carrier Grade NAT setup with two relays and a single signaling server
 *
 * Hosts:
 *  - 2x Relay nodes (Coturn STUN/TURN server)
 *  - 1x Signaling node (GRPC server)
 *  - 3x NAT routers
 *  - 2x WAN switches
 *  - 2x LAN switches
 *  - 2x wice Agent nodes
 *
 *             ┌──────┐
 *             │  s1  │            Signaling
 *             └──┬───┘
 *  ┌──────┐      │       ┌──────┐
 *  │  r1  │      │       │  r2  │ Relays
 *  └──┬───┘      │       └──┬───┘
 *  ┌──┴───┐      │       ┌──┴───┐
 *  │ sw1  ├──────┘       │  sw2 │ WAN Switches
 *  └──┬─┬─┘              └─┬─┬──┘
 *     │ └───┐          ┌───┘ │
 *  ┌──┴───┐ │ ┌──────┐ │ ┌───┴──┐
 *  │ nat1 │ └─┤ nat3 ├─┘ │ nat2 │ NAT Routers
 *  └──┬───┘   └──────┘   └───┬──┘
 *  ┌──┴───┐              ┌───┴──┐
 *  │ lsw1 │              │ lsw2 │ LAN Switches
 *  └──┬───┘              └───┬──┘
 *  ┌──┴───┐              ┌───┴──┐
 *  │  n1  │              │  n2  │ wice Agents
 *  └──────┘              └──────┘
 */
var _ = Context("nat double", Serial, func() {
	var n Network

	BeforeEach(func() {
		n.Init()
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

		sw2, err := nw.AddSwitch("sw2")
		Expect(err).To(Succeed(), "Failed to create switch: %s", err)

		lsw1, err := nw.AddSwitch("lws1")
		Expect(err).To(Succeed(), "Failed to create switch: %s", err)

		lsw2, err := nw.AddSwitch("lws2")
		Expect(err).To(Succeed(), "Failed to create switch: %s", err)

		By("Initializing relay node")

		r1, err := nodes.NewCoturnNode(nw, "r1",
			gopt.Interface("eth0", sw1,
				gopt.AddressIP("10.10.0.1/16"),
				gopt.AddressIP("fc:10::1/64"),
			),
		)
		Expect(err).To(Succeed(), "Failed to start relay: %s", err)

		r2, err := nodes.NewCoturnNode(nw, "r2",
			gopt.Interface("eth0", sw2,
				gopt.AddressIP("10.11.0.1/16"),
				gopt.AddressIP("fc:11::1/64"),
			),
		)
		Expect(err).To(Succeed(), "Failed to start relay: %s", err)

		By("Initializing signaling node")

		s1, err := nodes.NewGrpcSignalingNode(nw, "s1",
			gopt.Interface("eth0", sw1,
				gopt.AddressIP("10.10.0.2/16"),
				gopt.AddressIP("fc:10::2/64"),
			),
		)
		Expect(err).To(Succeed(), "Failed to create signaling node: %s", err)

		By("Initializing CGNAT node")

		_, err = nw.AddNAT("nat3",
			gopt.Interface("eth-nb", sw1,
				gopt.NorthBound,
				gopt.AddressIP("10.10.0.4/16"),
				gopt.AddressIP("fc:10::4/64"),
			),
			gopt.Interface("eth-sb", sw2,
				gopt.SouthBound,
				gopt.AddressIP("10.11.0.4/24"),
				gopt.AddressIP("fc:11::4/64"),
			),
		)
		Expect(err).To(Succeed(), "Failed to add NAT node: %s", err)

		By("Initializing agent nodes")

		/// Left: n1

		opts := gopt.Customize(n.AgentOptions,
			gopt.Interface("eth0", lsw1,
				gopt.AddressIP("10.1.0.2/24"),
				gopt.AddressIP("fc:1::2/64"),
			),
			gopt.DefaultGatewayIP("10.1.0.1"),
			gopt.DefaultGatewayIP("fc:1::1"),
			wopt.Interface("wg0",
				wopt.AddressIP("172.16.0.1/16"),
				wopt.PeerFromNames("n2", "wg0",
					wopt.AllowedIPStr("172.16.0.2/32"),
				),
			),
		)

		n1, err := nodes.NewAgent(nw, "n1", opts...)
		Expect(err).To(Succeed(), "ailed to created nodes: %s", err)

		_, err = nw.AddNAT("nat1",
			gopt.Interface("eth-nb", sw1,
				gopt.NorthBound,
				gopt.AddressIP("10.10.0.3/16"),
				gopt.AddressIP("fc:10::3/64"),
			),
			gopt.Interface("eth-sb", lsw1,
				gopt.SouthBound,
				gopt.AddressIP("10.1.0.1/24"),
				gopt.AddressIP("fc:1::1/64"),
			),
		)
		Expect(err).To(Succeed(), "Failed to add NAT node: %s", err)

		/// Right: n2

		opts = gopt.Customize(n.AgentOptions,
			gopt.Interface("eth0", lsw2,
				gopt.AddressIP("10.1.0.2/24"),
				gopt.AddressIP("fc:1::2/64"),
			),
			gopt.DefaultGatewayIP("10.1.0.1"),
			gopt.DefaultGatewayIP("fc:1::1"),
			wopt.Interface("wg0",
				wopt.AddressIP("172.16.0.2/16"),
				wopt.PeerFromNames("n1", "wg0",
					wopt.AllowedIPStr("172.16.0.1/32"),
				),
			),
		)

		n2, err := nodes.NewAgent(nw, "n2", opts...)
		Expect(err).To(Succeed(), "Failed to create nodes: %s", err)

		_, err = nw.AddNAT("nat2",
			gopt.DefaultGatewayIP("10.11.0.4"),
			gopt.DefaultGatewayIP("fc:11::4"),
			gopt.Interface("eth-nb", sw2,
				gopt.NorthBound,
				gopt.AddressIP("10.11.0.3/16"),
				gopt.AddressIP("fc:11::3/64"),
			),
			gopt.Interface("eth-sb", lsw2,
				gopt.SouthBound,
				gopt.AddressIP("10.1.0.1/24"),
				gopt.AddressIP("fc:1::1/64"),
			),
		)
		Expect(err).To(Succeed(), "Failed to add NAT node: %s", err)

		n.Network = nw
		n.AgentNodes = nodes.AgentList{n1, n2}
		n.RelayNodes = nodes.RelayList{r1, r2}
		n.SignalingNodes = nodes.SignalingList{s1}

		n.Start()
	})

	n.ConnectivityTests()
})
