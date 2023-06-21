// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	g "github.com/stv0g/gont/v2/pkg"
	gopt "github.com/stv0g/gont/v2/pkg/options"

	"github.com/stv0g/cunicu/pkg/wg"
	"github.com/stv0g/cunicu/test/e2e/nodes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("single: A single node to test RPC and watcher", Pending, func() {
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

		By("Initializing agent node")

		n1, err := nodes.NewAgent(nw, "n1", n.AgentOptions...)
		Expect(err).To(Succeed(), "Failed to create agent node: %s", err)

		By("Starting network")

		n.Network = nw
		n.AgentNodes = nodes.AgentList{n1}
	})

	Context("create: Create a new interface", func() {
		Context("kernel: Kernel-space", func() {
		})

		Context("userspace: User-space", func() {
		})
	})

	Context("watcher: Watch for changes of WireGuard interfaces and peers", Ordered, func() {
		It("detects a new interface", func() {
		})

		It("detects a change of the interface", func() {
		})

		It("detects a new peer", func() {
		})

		It("detects a change of the peer", func() {
		})

		It("detects the removal of the peer", func() {
		})

		It("detects the removal of the interface", func() {
		})
	})

	Context("autocfg: Auto-configuration of missing interface parameters", Pending, func() {
	})

	Context("cfgsync: Config file synchronization", Pending, func() {
	})

	Context("hooks: Hook execution", Pending, func() {
	})

	Context("hsync: /etc/hosts synchronization", Pending, func() {
	})

	Context("rtsync: Route synchronization", Pending, func() {
	})
})
