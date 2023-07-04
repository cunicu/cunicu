// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"

	g "github.com/stv0g/gont/v2/pkg"
	gopt "github.com/stv0g/gont/v2/pkg/options"

	"github.com/stv0g/cunicu/test/e2e/nodes"
	wopt "github.com/stv0g/cunicu/test/e2e/nodes/options/wg"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

/* Typical wide-area NAT setup
 *
 * Hosts:
 *  - 1x Relay node       [r1] (Coturn STUN/TURN server)
 *  - 1x Signaling node   [s1] (GRPC server)
 *  - 2x NAT routers      [nat?]
 *  - 1x WAN switch       [wan?]
 *  - 2x LAN switches     [lan?]
 *  - 2x cunicu Agent nodes [n?]git restore --staged
 *
 *     ┌──────┐   ┌──────┐
 *     │  r1  │   │  s1  │
 *     └───┬──┘   └──┬───┘
 *         └───┐  ┌──┘
 *           ┌─┴──┴─┐
 *           │ wan1 │ WAN Switch
 *           └┬────┬┘
 *      ┌─────┘    └────┐
 *  ┌───┴──┐        ┌───┴──┐
 *  │ nat1 │        │ nat2 │
 *  └───┬──┘        └───┬──┘
 *  ┌───┴──┐        ┌───┴──┐
 *  │ lan1 │        │ lan2 │
 *  └───┬──┘        └─┬──┬─┘
 *      │           ┌─┘  └─┐
 *  ┌───┴──┐   ┌────┴─┐  ┌─┴────┐
 *  │  n1  │   │  n2  │  │ (n3) │
 *  └──────┘   └──────┘  └──────┘
 */
var _ = Context("nat simple: Simple home-router NAT setup", Pending, func() {
	var (
		err error

		n    Network
		nw   *g.Network
		lan2 *g.Switch
	)

	BeforeEach(func() {
		n.Init()
	})

	AfterEach(func() {
		n.Close()
	})

	AddAgent := func(i int, lan *g.Switch) *nodes.Agent {
		a, err := nodes.NewAgent(nw, fmt.Sprintf("n%d", i),
			gopt.DefaultGatewayIP("10.1.0.254"),
			gopt.DefaultGatewayIP("fc:1::254"),
			g.NewInterface("eth0", lan,
				gopt.AddressIP("10.1.0.%d/24", i),
				gopt.AddressIP("fc:1::%d/64", i),
			),
			wopt.Interface("wg0",
				wopt.FullMeshPeers,
				wopt.AddressIP("172.16.0.%d/16", i),
			),
		)
		Expect(err).To(Succeed(), "Failed to created nodes: %s", err)

		n.AgentNodes = append(n.AgentNodes, a)

		return a
	}

	JustBeforeEach(func() {
		By("Initializing core network")
		nw, err = g.NewNetwork(n.Name, n.NetworkOptions...)

		Expect(err).To(Succeed(), "Failed to create network: %s", err)

		wan1, err := nw.AddSwitch("wan1")
		Expect(err).To(Succeed(), "Failed to create switch: %s", err)

		By("Initializing relay node")

		r1, err := nodes.NewCoturnNode(nw, "r1",
			g.NewInterface("eth0", wan1,
				gopt.AddressIP("10.0.0.1/16"),
				gopt.AddressIP("fc::1/64"),
			),
		)
		Expect(err).To(Succeed(), "Failed to start relay: %s", err)

		By("Initializing signaling node")

		s1, err := nodes.NewGrpcSignalingNode(nw, "s1",
			g.NewInterface("eth0", wan1,
				gopt.AddressIP("10.0.0.2/16"),
				gopt.AddressIP("fc::2/64"),
			),
		)
		Expect(err).To(Succeed(), "Failed to create signaling node: %s", err)

		By("Initializing agent nodes")

		AddLAN := func(i int) *g.Switch {
			// LAN switch
			lan, err := nw.AddSwitch(fmt.Sprintf("lan%d", i))
			Expect(err).To(Succeed(), "Failed to add LAN switch: %s", err)

			// NAT router
			_, err = nw.AddNAT(fmt.Sprintf("nat%d", i),
				g.NewInterface("eth-nb", wan1,
					gopt.NorthBound,
					gopt.AddressIP("10.0.1.%d/16", i),
					gopt.AddressIP("fc::1:%d/64", i),
				),
				g.NewInterface("eth-sb", lan,
					gopt.SouthBound,
					gopt.AddressIP("10.1.0.254/24"),
					gopt.AddressIP("fc:1::254/64"),
				),
			)
			Expect(err).To(Succeed(), "Failed to created nodes: %s", err)

			AddAgent(i, lan)

			return lan
		}

		AddLAN(1)
		lan2 = AddLAN(2)

		n.Network = nw
		n.RelayNodes = nodes.RelayList{r1}
		n.SignalingNodes = nodes.SignalingList{s1}
	})

	Context("2-nodes: Two agents connected to lan1 & lan2", func() {
		JustBeforeEach(func() {
			n.Start()
		})

		n.ConnectivityTests()
	})

	Context("3-nodes: Additional agent connected to lan2", func() {
		JustBeforeEach(func() {
			AddAgent(3, lan2)

			n.Start()
		})

		n.ConnectivityTests()
	})
})
