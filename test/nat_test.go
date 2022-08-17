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

/* Typical wide-area NAT setup
 *
 * Hosts:
 *  - 1x Relay node (Coturn STUN/TURN server)
 *  - 1x Signaling node (GRPC server)
 *  - 2x NAT routers
 *  - 1x WAN switch
 *  - 2x LAN switches
 *  - 2x wice Agent nodes
 *
 *        ┌────┐   ┌────┐
 *  Relay │ r1 │   │ s1 │ Signaling
 *        └──┬─┘   └─┬──┘
 *           └─┐   ┌─┘
 *            ┌┴───┴┐
 *            │ sw1 │ WAN Switch
 *            └┬───┬┘
 *          ┌──┘   └──┐
 *      ┌───┴──┐   ┌──┴───┐
 *      │ nat1 │   │ nat2 │ NAT Routers
 *      └───┬──┘   └──┬───┘
 *      ┌───┴──┐   ┌──┴───┐
 *      │ lsw1 │   │ lsw2 │ LAN Switches
 *      └───┬──┘   └──┬───┘
 *      ┌───┴──┐   ┌──┴───┐
 *      │  n1  │   │  n2  │ wice Agents
 *      └──────┘   └──────┘
 */
var _ = Context("nat simple", Serial, func() {
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

		/// n1

		lsw1, err := nw.AddSwitch("lsw1")
		Expect(err).To(Succeed(), "Failed to add LAN switch: %s", err)

		opts := gopt.Customize(n.AgentOptions,
			gopt.Interface("eth0", lsw1,
				gopt.AddressIP("10.1.0.2/24"),
				gopt.AddressIP("fc:1::2/64"),
			),
			gopt.DefaultGatewayIP("10.1.0.1"),
			gopt.DefaultGatewayIP("fc:1::1"),
			wopt.Interface("wg0",
				wopt.AddressIP("172.16.0.1/16"),
				wopt.FullMeshPeers,
			),
		)

		n1, err := nodes.NewAgent(nw, "n1", opts...)
		Expect(err).To(Succeed(), "Failed to created nodes: %s", err)

		_, err = nw.AddNAT("nat1",
			gopt.Interface("eth-nb", sw1,
				gopt.NorthBound,
				gopt.AddressIP("10.0.1.1/16"),
				gopt.AddressIP("fc::1:1/64"),
			),
			gopt.Interface("eth-sb", lsw1,
				gopt.SouthBound,
				gopt.AddressIP("10.1.0.1/24"),
				gopt.AddressIP("fc:1::1/64"),
			),
		)
		Expect(err).To(Succeed(), "Failed to created nodes: %s", err)

		/// n2

		lsw2, err := nw.AddSwitch("lsw2")
		Expect(err).To(Succeed(), "Failed to add LAN switch: %s", err)

		opts = gopt.Customize(n.AgentOptions,
			gopt.Interface("eth0", lsw2,
				gopt.AddressIP("10.1.0.2/24"),
				gopt.AddressIP("fc:1::2/64"),
			),
			gopt.DefaultGatewayIP("10.1.0.1"),
			gopt.DefaultGatewayIP("fc:1::1"),
			wopt.Interface("wg0",
				wopt.AddressIP("172.16.0.2/16"),
				wopt.FullMeshPeers,
			),
		)

		n2, err := nodes.NewAgent(nw, "n2", opts...)
		Expect(err).To(Succeed(), "Failed to created nodes: %s", err)

		_, err = nw.AddNAT("nat2",
			gopt.Interface("eth-nb", sw1,
				gopt.NorthBound,
				gopt.AddressIP("10.0.1.2/16"),
				gopt.AddressIP("fc::1:2/64"),
			),
			gopt.Interface("eth-sb", lsw2,
				gopt.SouthBound,
				gopt.AddressIP("10.1.0.1/24"),
				gopt.AddressIP("fc:1::1/64"),
			),
		)
		Expect(err).To(Succeed(), "Failed to created nodes: %s", err)

		n.Network = nw
		n.AgentNodes = nodes.AgentList{n1, n2}
		n.RelayNodes = nodes.RelayList{r1}
		n.SignalingNodes = nodes.SignalingList{s1}

		n.Start()
	})

	n.ConnectivityTests()
})
