//go:build linux

package test_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	g "github.com/stv0g/gont/pkg"
	gopt "github.com/stv0g/gont/pkg/options"
	"riasc.eu/wice/test/nodes"
	opt "riasc.eu/wice/test/nodes/options"
	wopt "riasc.eu/wice/test/nodes/options/wg"
)

/* Simple local-area switched topology for scalability tests
 *
 * Hosts:
 *  - 1x Relay node (Coturn STUN/TURN server)
 *  - 1x Signaling node (GRPC server)
 *  - 1x Switch
 *  - Yx  wice Agent nodes
 *
 *        ┌────┐            ┌────┐
 *  Relay │ r1 │            │ s1 │ Signaling
 *        └──┬─┘            └─┬──┘
 *           └─────┐   ┌──────┘
 *                ┌┴───┴┐
 *                │ sw1 │ Switch
 *                └┬─┬─┬┘
 *           ┌─────┘ │ └───────┐
 *        ┌──┴─┐  ┌──┴─┐     ┌─┴──┐
 *        │ n1 │  │ n2 │ ... │ nY │ wice Agents
 *        └────┘  └────┘     └────┘
 */
var _ = Context("scaling", Serial, func() {
	var (
		n Network

		NumAgents int
	)

	BeforeEach(func() {
		n.Init()

		n.AgentOptions = append(n.AgentOptions,
			opt.ExtraArgs{
				"--host-sync=false",
				"--ice-max-binding-requests=500",
				// "--ice-check-interval=1s",
				"--ice-failed-timeout=10s",
				"--ice-candidate-type=host",
				"--ice-candidate-type=srflx",
			},
		)
		NumAgents = 4
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

		for i := 1; i <= NumAgents; i++ {
			node, err := nodes.NewAgent(nw, fmt.Sprintf("n%d", i),
				gopt.Customize(n.AgentOptions,
					gopt.Interface("eth0", sw1,
						gopt.AddressIPv4(10, 0, 1, byte(i), 16),
						gopt.AddressIP(fmt.Sprintf("fc::1:%d/64", i)),
					),
					wopt.Interface("wg0",
						wopt.AddressIPv4(172, 16, 0, byte(i), 16),
						wopt.FullMeshPeers,
					),
				)...,
			)
			Expect(err).To(Succeed(), "Failed to create node: %s", err)

			n.AgentNodes = append(n.AgentNodes, node)
		}

		n.Network = nw
		n.RelayNodes = nodes.RelayList{r1}
		n.SignalingNodes = nodes.SignalingList{s1}

		n.Start()
	})

	n.ConnectivityTests()
})
