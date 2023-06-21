// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"context"
	"fmt"
	"time"

	g "github.com/stv0g/gont/v2/pkg"
	gopt "github.com/stv0g/gont/v2/pkg/options"
	copt "github.com/stv0g/gont/v2/pkg/options/cmd"

	"github.com/stv0g/cunicu/pkg/crypto"
	netx "github.com/stv0g/cunicu/pkg/net"
	"github.com/stv0g/cunicu/pkg/proto"
	"github.com/stv0g/cunicu/pkg/wg"
	"github.com/stv0g/cunicu/test/e2e/nodes"
	opt "github.com/stv0g/cunicu/test/e2e/nodes/options"
	wopt "github.com/stv0g/cunicu/test/e2e/nodes/options/wg"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

/* Simple local-area switched topology with 2 agents
 *
 *  - 1x Signaling node    [s1] (GRPC server)
 *  - 1x Switch            [sw1]
 *  - 2x  cunicu Agent nodes [n?]
 *
 *         Signaling
 *          ┌─────┐
 *          │  s1 │
 *          └──┬──┘
 *             │
 *          ┌──┴──┐
 *          │ sw1 │ Switch
 *          └┬───┬┘
 *       ┌───┘   └───┐
 *    ┌──┴─┐       ┌─┴──┐
 *    │ n1 │       │ n2 │
 *    └────┘       └────┘
 *         cunicu Agents
 */
var _ = Context("restart: Restart ICE agents", func() {
	var (
		err error
		n   Network
		nw  *g.Network

		s1     *nodes.GrpcSignalingNode
		n1, n2 *nodes.Agent
	)

	BeforeEach(OncePerOrdered, func() {
		n.Init()

		n.AgentOptions = append(n.AgentOptions,
			gopt.EmptyDir(wg.ConfigPath),
			gopt.EmptyDir(wg.SocketPath),
			opt.ExtraArgs{"--ice-candidate-type", "host"},
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

		By("Initializing signaling node")

		s1, err = nodes.NewGrpcSignalingNode(nw, "s1",
			g.NewInterface("eth0", sw1,
				gopt.AddressIP("10.0.0.2/16"),
				gopt.AddressIP("fc::2/64"),
			),
		)
		Expect(err).To(Succeed(), "Failed to create signaling node: %s", err)

		By("Initializing agent nodes")

		AddAgent := func(i int) *nodes.Agent {
			a, err := nodes.NewAgent(nw, fmt.Sprintf("n%d", i),
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
			Expect(err).To(Succeed(), "Failed to create agent node: %s", err)

			n.AgentNodes = append(n.AgentNodes, a)

			return a
		}

		n1 = AddAgent(1)
		n2 = AddAgent(2)

		By("Starting network")

		n.Network = nw
		n.SignalingNodes = nodes.SignalingList{s1}

		n.Start()
	})

	RestartTest := func(restart func(gap time.Duration)) {
		var gap time.Duration

		ConnectivityTestCycle := func() {
			n.ConnectivityTests()

			It("Triggering restart", func() {
				restart(gap)

				time.Sleep(gap)
			})

			n.ConnectivityTests()
		}

		Context("quick: Waiting 3 seconds", Ordered, func() {
			BeforeEach(func() {
				gap = 3 * time.Second
			})

			ConnectivityTestCycle()
		})

		Context("slow: Waiting 10 seconds to trigger an ICE disconnect", Ordered, func() {
			BeforeEach(func() {
				gap = 10 * time.Second // > ICE failed/disconnected timeout (5s)
			})

			ConnectivityTestCycle()
		})
	}

	Context("agent: Restart agent", func() {
		RestartTest(func(gap time.Duration) {
			By("Stopping first agent")

			err = n1.Stop()
			Expect(err).To(Succeed(), "Failed to stop first agent: %s", err)

			By("Waiting some time")

			time.Sleep(gap)

			By("Re-starting first agent again")

			err = n1.Start("", n.BasePath, n.AgentArgs()...)
			Expect(err).To(Succeed(), "Failed to restart first agent: %s", err)
		})
	})

	Context("addresses: Change uplink IP address", Pending, func() {
		RestartTest(func(gap time.Duration) {
			i := n1.Interface("eth0")
			Expect(i).NotTo(BeNil(), "Failed to find agent interface")

			By("Deleting old addresses from agent interface")

			for _, a := range i.Addresses {
				a := a
				err = i.DeleteAddress(&a)
				Expect(err).To(Succeed(), "Failed to remove IP address '%s': %s", a, err)
			}

			By("Waiting some time")

			time.Sleep(gap)

			By("Assigning new addresses to agent interface")

			for _, a := range i.Addresses {
				ao := a
				ao.IP = netx.OffsetIP(ao.IP, 128)

				err = i.AddAddress(&ao)
				Expect(err).To(Succeed(), "Failed to add IP address '%s': %s", a, err)
			}

			n1.Run("ip", "a", copt.Combined(GinkgoWriter)) //nolint:errcheck
			n1.Run("wg", copt.Combined(GinkgoWriter))      //nolint:errcheck
			n2.Run("wg", copt.Combined(GinkgoWriter))      //nolint:errcheck
		})
	})

	Context("link: Bring uplink down and up", Pending, func() {
		RestartTest(func(gap time.Duration) {
			i := n1.Interface("eth0")
			Expect(i).NotTo(BeNil(), "Failed to find agent interface")

			By("Bringing interface of first agent down")

			err = i.SetDown()
			Expect(err).To(Succeed(), "Failed to bring interface down: %s", err)

			By("Waiting some time")

			time.Sleep(gap)

			By("Bringing interface of first agent back up")

			err = i.SetUp()
			Expect(err).To(Succeed(), "Failed to bring interface back up: %s", err)
		})
	})

	Context("peer-rpc: Restart peer via RPC", func() {
		RestartTest(func(gap time.Duration) {
			ctx := context.Background()

			By("Initiating restart via RPC")

			i := n1.WireGuardInterfaces[0]
			pk := (*crypto.Key)(&i.Peers[0].PublicKey)
			err = n1.Client.RestartPeer(ctx, i.Name, pk)
			Expect(err).To(Succeed(), "Failed to restart peer: %s", err)
		})
	})

	Context("agent-rpc: Restart agent via RPC", Pending, func() {
		RestartTest(func(gap time.Duration) {
			ctx := context.Background()

			By("Initiating agent restart via RPC")

			_, err = n1.Client.Restart(ctx, &proto.Empty{})
			Expect(err).To(Succeed(), "Failed to restart peer: %s", err)
		})
	})

	Context("signaling: Restart signaling server", Pending, func() {
		RestartTest(func(gap time.Duration) {
			ctx := context.Background()

			By("Stopping signaling server")

			err = s1.Stop()
			Expect(err).To(Succeed(), "Failed to stop signaling server")

			By("Restarting peer")

			i := n1.WireGuardInterfaces[0]
			pk := (*crypto.Key)(&i.Peers[0].PublicKey)
			err = n1.Client.RestartPeer(ctx, i.Name, pk)
			Expect(err).To(Succeed(), "Failed to restart peer: %s", err)

			By("Waiting some time")

			time.Sleep(gap)

			By("Starting signaling server again")

			err = s1.Start("", n.BasePath)
			Expect(err).To(Succeed(), "Failed to restart signaling server: %s", err)
		})
	})

	Context("private-key: Trigger recreation of peer by changing private-key of interface", func() {
		RestartTest(func(gap time.Duration) {
			By("Changing private-key")

			// i.Configuring
		})
	})
})
