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

/* Carrier Grade NAT setup with two relays and a single signaling server
 *
 * Hosts:
 *  - 1x Signaling node     [s1]    (GRPC server)
 *  - 2x Relay nodes        [nat?]  (Coturn STUN/TURN server)
 *  - 3x NAT routers        [nat?]
 *  - 2x WAN switches       [wan?]
 *  - 2x LAN switches       [lan?]
 *  - 2-5x cunicu Agent nodes [n?]
 *
 *
 *             ┌──────┐   ┌──────┐   ┌──────┐
 *             │  r1  │   │  s1  │   │  r2  │
 *             └──┬───┘   └──┬───┘   └───┬──┘
 *                │ ┌────────┘           │
 *  ┌──────┐   ┌──┴─┴─┐   ┌──────┐   ┌───┴──┐   ┌──────┐
 *  │ (n5) ├───┤ wan1 ├───┤ nat3 ├───┤ wan2 ├───┤ (n4) │
 *  └──────┘   └──┬───┘   └──────┘   └───┬──┘   └──────┘
 *             ┌──┴───┐              ┌───┴──┐
 *             │ nat1 │              │ nat2 │
 *             └──┬───┘              └───┬──┘
 *             ┌──┴───┐              ┌───┴──┐   ┌──────┐
 *             │ lan1 │              │ lan2 ├───┤ (n3) │
 *             └──┬───┘              └───┬──┘   └──────┘
 *             ┌──┴───┐              ┌───┴──┐
 *             │  n1  │              │  n2  │
 *             └──────┘              └──────┘
 */
var _ = Context("nat double: Carrier Grade NAT setup with two relays and a single signaling server", Pending, func() {
	var (
		err error

		n Network

		nw         *g.Network
		wan1, wan2 *g.Switch
		lan1, lan2 *g.Switch
	)

	BeforeEach(func() {
		n.Init()
	})

	AfterEach(func() {
		n.Close()
	})

	AddAgent := func(i int, sw *g.Switch) *nodes.Agent {
		ifOpts := []g.Option{sw}

		switch {
		case i <= 3: // lan1, lan2
			ifOpts = append(ifOpts,
				gopt.AddressIP("10.1.0.%d/24", i),
				gopt.AddressIP("fc:1::%d/64", i),
			)
		case i == 4: // wan2
			ifOpts = append(ifOpts,
				gopt.AddressIP("10.11.0.5/24"),
				gopt.AddressIP("fc:11::5/64"),
			)
		case i == 5: // wan1
			ifOpts = append(ifOpts,
				gopt.AddressIP("10.10.0.5/24"),
				gopt.AddressIP("fc:10::5/64"),
			)
		}

		opts := gopt.Customize[g.Option](n.AgentOptions,
			g.NewInterface("eth0", ifOpts...),
			wopt.Interface("wg0",
				wopt.AddressIP("172.16.0.%d/16", i),
				wopt.FullMeshPeers,
			),
		)

		switch {
		case i <= 3: // lan1, lan2
			opts = append(opts,
				gopt.DefaultGatewayIP("10.1.0.254"),
				gopt.DefaultGatewayIP("fc:1::ff"),
			)
		case i == 4: // wan2
			opts = append(opts,
				gopt.DefaultGatewayIP("10.11.0.4"),
				gopt.DefaultGatewayIP("fc:11::4"),
			)
		}

		a, err := nodes.NewAgent(nw, fmt.Sprintf("n%d", i), opts...)
		Expect(err).To(Succeed(), "Failed to created nodes: %s", err)

		n.AgentNodes = append(n.AgentNodes, a)

		return a
	}

	AddLAN := func(i int, sw *g.Switch) *g.Switch {
		// LAN Switch
		lan, err := nw.AddSwitch(fmt.Sprintf("lan%d", i))
		Expect(err).To(Succeed(), "Failed to create switch: %s", err)

		nbIfOpts := []g.Option{sw, gopt.NorthBound}

		switch {
		case i == 1: // wan1
			nbIfOpts = append(nbIfOpts,
				gopt.AddressIP("10.10.0.3/16"),
				gopt.AddressIP("fc:10::3/64"),
			)
		case i == 2: // wan2
			nbIfOpts = append(nbIfOpts,
				gopt.AddressIP("10.11.0.3/16"),
				gopt.AddressIP("fc:11::3/64"),
			)
		}

		opts := []g.Option{
			g.NewInterface("eth-nb", nbIfOpts...),
			g.NewInterface("eth-sb", lan,
				gopt.SouthBound,
				gopt.AddressIP("10.1.0.254/24"),
				gopt.AddressIP("fc:1::ff/64"),
			),
		}

		switch {
		case i == 2: // wan2
			opts = append(opts,
				gopt.DefaultGatewayIP("10.11.0.4"),
				gopt.DefaultGatewayIP("fc:11::4"),
			)
		}

		// NAT router
		_, err = nw.AddNAT(fmt.Sprintf("nat%d", i), opts...)
		Expect(err).To(Succeed(), "Failed to add NAT node: %s", err)

		return lan
	}

	JustBeforeEach(func() {
		By("Initializing core network")

		nw, err = g.NewNetwork(n.Name, n.NetworkOptions...)
		Expect(err).To(Succeed(), "Failed to create network: %s", err)

		wan1, err = nw.AddSwitch("wan1")
		Expect(err).To(Succeed(), "Failed to create switch: %s", err)

		wan2, err = nw.AddSwitch("wan2")
		Expect(err).To(Succeed(), "Failed to create switch: %s", err)

		By("Initializing relay node")

		r1, err := nodes.NewCoturnNode(nw, "r1",
			g.NewInterface("eth0", wan1,
				gopt.AddressIP("10.10.0.1/16"),
				gopt.AddressIP("fc:10::1/64"),
			),
		)
		Expect(err).To(Succeed(), "Failed to start relay: %s", err)

		r2, err := nodes.NewCoturnNode(nw, "r2",
			g.NewInterface("eth0", wan2,
				gopt.AddressIP("10.11.0.1/16"),
				gopt.AddressIP("fc:11::1/64"),
			),
		)
		Expect(err).To(Succeed(), "Failed to start relay: %s", err)

		By("Initializing signaling node")

		s1, err := nodes.NewGrpcSignalingNode(nw, "s1",
			g.NewInterface("eth0", wan1,
				gopt.AddressIP("10.10.0.2/16"),
				gopt.AddressIP("fc:10::2/64"),
			),
		)
		Expect(err).To(Succeed(), "Failed to create signaling node: %s", err)

		By("Initializing CGNAT node")

		_, err = nw.AddNAT("nat3",
			g.NewInterface("eth-nb", wan1,
				gopt.NorthBound,
				gopt.AddressIP("10.10.0.4/16"),
				gopt.AddressIP("fc:10::4/64"),
			),
			g.NewInterface("eth-sb", wan2,
				gopt.SouthBound,
				gopt.AddressIP("10.11.0.4/24"),
				gopt.AddressIP("fc:11::4/64"),
			),
		)
		Expect(err).To(Succeed(), "Failed to add NAT node: %s", err)

		By("Initializing agent nodes")

		lan1 = AddLAN(1, wan1)
		lan2 = AddLAN(2, wan2)

		AddAgent(1, lan1)
		AddAgent(2, lan2)

		n.Network = nw
		n.RelayNodes = nodes.RelayList{r1, r2}
		n.SignalingNodes = nodes.SignalingList{s1}
	})

	Context("2-nodes: Two agents connected to lan1 and lan2", func() {
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

	Context("4-nodes: Additional agent connected to wan2", func() {
		JustBeforeEach(func() {
			AddAgent(3, lan2)
			AddAgent(4, wan2)

			n.Start()
		})

		n.ConnectivityTests()
	})

	Context("5-nodes: Additional agent connected to wan1", func() {
		JustBeforeEach(func() {
			AddAgent(3, lan2)
			AddAgent(4, wan2)
			AddAgent(5, wan1)

			n.Start()
		})

		n.ConnectivityTests()
	})
})
