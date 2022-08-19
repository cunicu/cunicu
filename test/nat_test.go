//go:build linux

package test_test

import (
	"fmt"

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
 *                    ┌────┐   ┌────┐
 *              Relay │ r1 │   │ s1 │ Signaling
 *                    └──┬─┘   └─┬──┘
 *                       └─┐   ┌─┘
 *                        ┌┴───┴┐
 *                        │ sw1 │ WAN Switch
 *                        └┬───┬┘
 *                   ┌─────┘   └─────┐
 *               ┌───┴──┐        ┌───┴──┐
 *               │ nat1 │        │ nat2 │ NAT Routers
 *               └───┬──┘        └───┬──┘
 *               ┌───┴──┐        ┌───┴──┐
 *  LAN Switches │ lsw1 │        │ lsw2 │
 *               └───┬──┘        └─┬──┬─┘
 *                   │           ┌─┘  └─┐
 *               ┌───┴──┐   ┌────┴─┐  ┌─┴────┐
 *               │  n1  │   │  n2  │  │ (n3) │  wice Agents
 *               └──────┘   └──────┘  └──────┘
 */
var _ = Context("nat simple", Serial, func() {
	var n Network
	var lsw2 *g.Switch

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

		CreateAgent := func(i int) *nodes.Agent {
			// LAN switch
			lsw, err := nw.AddSwitch(fmt.Sprintf("lsw%d", i))
			Expect(err).To(Succeed(), "Failed to add LAN switch: %s", err)

			if i == 2 {
				lsw2 = lsw
			}

			// NAT router
			_, err = nw.AddNAT(fmt.Sprintf("nat%d", i),
				gopt.Interface("eth-nb", sw1,
					gopt.NorthBound,
					gopt.AddressIP("10.0.1.%d/16", i),
					gopt.AddressIP("fc::1:%d/64", i),
				),
				gopt.Interface("eth-sb", lsw,
					gopt.SouthBound,
					gopt.AddressIP("10.1.0.1/24"),
					gopt.AddressIP("fc:1::1/64"),
				),
			)
			Expect(err).To(Succeed(), "Failed to created nodes: %s", err)

			// Agent node
			n, err := nodes.NewAgent(nw, fmt.Sprintf("n%d", i),
				gopt.Customize(n.AgentOptions,
					gopt.Interface("eth0", lsw,
						gopt.AddressIP("10.1.0.2/24"),
						gopt.AddressIP("fc:1::2/64"),
					),
					gopt.DefaultGatewayIP("10.1.0.1"),
					gopt.DefaultGatewayIP("fc:1::1"),
					wopt.Interface("wg0",
						wopt.AddressIP("172.16.0.%d/16", i),
						wopt.FullMeshPeers,
					),
				)...,
			)
			Expect(err).To(Succeed(), "Failed to created nodes: %s", err)

			return n
		}

		n.Network = nw
		n.AgentNodes = nodes.AgentList{
			CreateAgent(1),
			CreateAgent(2),
		}
		n.RelayNodes = nodes.RelayList{r1}
		n.SignalingNodes = nodes.SignalingList{s1}
	})

	Context("without-n3", func() {
		JustBeforeEach(func() {
			n.Start()
		})

		n.ConnectivityTests()
	})

	FContext("with-n3", func() {
		JustBeforeEach(func() {
			n3, err := nodes.NewAgent(n.Network, "n3",
				gopt.Customize(n.AgentOptions,
					gopt.Interface("eth0", lsw2,
						gopt.AddressIP("10.1.0.3/24"),
						gopt.AddressIP("fc:1::3/64"),
					),
					gopt.DefaultGatewayIP("10.1.0.1"),
					gopt.DefaultGatewayIP("fc:1::1"),
					wopt.Interface("wg0",
						wopt.AddressIP("172.16.0.3/16"),
						wopt.FullMeshPeers,
					),
				)...,
			)
			Expect(err).To(Succeed(), "Failed to created node: %s", err)

			n.AgentNodes = append(n.AgentNodes, n3)

			n.Start()
		})

		n.ConnectivityTests()
	})
})
