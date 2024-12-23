// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	g "cunicu.li/gont/v2/pkg"
	gopt "cunicu.li/gont/v2/pkg/options"
	copt "cunicu.li/gont/v2/pkg/options/capture"
	"github.com/knadh/koanf/v2"

	"cunicu.li/cunicu/pkg/log"
	osx "cunicu.li/cunicu/pkg/os"
	"cunicu.li/cunicu/pkg/tty"
	"cunicu.li/cunicu/test"
	"cunicu.li/cunicu/test/e2e/nodes"
	opt "cunicu.li/cunicu/test/e2e/nodes/options"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var logger *log.Logger

type Network struct {
	*g.Network

	Name string

	NetworkOptions            []g.NetworkOption
	AgentOptions              []g.Option
	WireGuardInterfaceOptions []g.Option

	BasePath string

	SignalingNodes nodes.SignalingList
	RelayNodes     nodes.RelayList
	AgentNodes     nodes.AgentList

	Tracer *HandshakeTracer
}

func (n *Network) Start() {
	By("Adding WireGuard peers")

	err := n.AgentNodes.ForEachInterfacePair(func(a, b *nodes.WireGuardInterface) error {
		if a.PeerSelector != nil && a.PeerSelector(a, b) {
			a.AddPeer(b)
		}
		return nil
	})
	Expect(err).To(Succeed(), "Failed to add WireGuard peers: %s", err)

	By("Configuring WireGuard interfaces")

	err = n.AgentNodes.ForEachAgent(func(a *nodes.Agent) error {
		return a.ConfigureWireGuardInterfaces()
	})
	Expect(err).To(Succeed(), "Failed to configure WireGuard interface: %s", err)

	if options.setup {
		Skip("Aborting test as only network setup has been requested")
	}

	if len(n.Captures) > 0 && n.Captures[0].LogKeys {
		n.StartHandshakeTracer()
	}

	By("Writing network hosts file")

	hfn := filepath.Join(n.BasePath, "hosts")
	hf, err := os.OpenFile(hfn, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	Expect(err).To(Succeed(), "Failed to open hosts file: %s", err)

	err = n.Network.WriteHostsFile(hf)
	Expect(err).To(Succeed(), "Failed to write hosts file: %s", err)

	err = hf.Close()
	Expect(err).To(Succeed(), "Failed to close hosts file: %s", err)

	By("Saving network nodes file")

	nfn := filepath.Join(n.BasePath, "nodes")
	nf, err := os.OpenFile(nfn, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	Expect(err).To(Succeed(), "Failed to open nodes file: %s", err)

	err = n.AgentNodes.ForEachInterface(func(i *nodes.WireGuardInterface) error {
		_, err := fmt.Fprintf(nf, "%s %s %s\n", i.Agent.Name(), i.Name, i.PrivateKey.PublicKey())
		return err
	})
	Expect(err).To(Succeed())

	err = nf.Close()
	Expect(err).To(Succeed(), "Failed to close nodes file: %s", err)

	By("Starting relay nodes")

	err = n.RelayNodes.Start(n.BasePath)
	Expect(err).To(Succeed(), "Failed to start relay: %s", err)

	By("Starting signaling nodes")

	err = n.SignalingNodes.Start(n.BasePath)
	Expect(err).To(Succeed(), "Failed to start signaling node: %s", err)

	By("Starting agent nodes")

	err = n.AgentNodes.Start(n.BasePath, n.AgentConfig())
	Expect(err).To(Succeed(), "Failed to start cunicu: %s", err)
}

func (n *Network) AgentConfig() nodes.AgentConfigOption {
	return opt.Config(func(k *koanf.Koanf) {
		var urls, backends []string

		for _, r := range n.RelayNodes {
			for _, u := range r.URLs() {
				urls = append(urls, u.String())
			}
		}

		for _, m := range n.SignalingNodes {
			u := m.URL()
			backends = append(backends, u.String())
		}

		k.Set("ice.urls", urls)     //nolint:errcheck
		k.Set("backends", backends) //nolint:errcheck
	})
}

func (n *Network) Close() {
	By("Stopping agent nodes")

	err := n.AgentNodes.Close()
	Expect(err).To(Succeed(), "Failed to close agent nodes; %s", err)

	By("Stopping signaling nodes")

	err = n.SignalingNodes.Close()
	Expect(err).To(Succeed(), "Failed to close signaling nodes; %s", err)

	By("Stopping relay nodes")

	err = n.RelayNodes.Close()
	Expect(err).To(Succeed(), "Failed to close relay nodes; %s", err)

	By("Stopping network")

	err = n.Network.Close()
	Expect(err).To(Succeed(), "Failed to close network; %s", err)

	n.StopHandshakeTracer()
	n.WriteSpecReport()

	GinkgoWriter.ClearTeeWriters()
}

func (n *Network) WriteSpecReport() {
	report := CurrentSpecReport()
	report.CapturedGinkgoWriterOutput = ""

	reportJSON, err := report.MarshalJSON()
	Expect(err).To(Succeed(), "Failed to marshal report: %s", err)

	reportJSON, err = tty.ReIndentJSON(reportJSON, "", "  ")
	Expect(err).To(Succeed(), "Failed to indent report: %s", err)

	reportFileName := filepath.Join(n.BasePath, "report.json")
	reportFile, err := os.OpenFile(reportFileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	Expect(err).To(Succeed(), "Failed to open report file: %s", err)

	_, err = reportFile.Write(reportJSON)
	Expect(err).To(Succeed(), "Failed to write report: %s", err)
}

func (n *Network) Init() {
	*n = Network{}

	cwd, err := os.Getwd()
	Expect(err).To(Succeed())

	n.Name = fmt.Sprintf("cunicu-%d", rand.Uint32()) //nolint:gosec
	n.BasePath = filepath.Join(SpecName()...)
	n.BasePath = filepath.Join(cwd, "logs", n.BasePath)

	logFilename := filepath.Join(n.BasePath, "test.log")
	pcapFilename := filepath.Join(n.BasePath, "capture.pcapng")

	By("Tweaking sysctls for large Gont networks")

	err = osx.SetSysctlMap(map[string]any{
		"net.ipv4.neigh.default.gc_thresh1": 10000,
		"net.ipv4.neigh.default.gc_thresh2": 15000,
		"net.ipv4.neigh.default.gc_thresh3": 20000,
		"net.ipv6.neigh.default.gc_thresh1": 10000,
		"net.ipv6.neigh.default.gc_thresh2": 15000,
		"net.ipv6.neigh.default.gc_thresh3": 20000,
		"net.core.rmem_max":                 32 << 20,
		"net.core.wmem_max":                 32 << 20,
		"net.core.rmem_default":             32 << 20,
	})
	Expect(err).To(Succeed(), "Failed to set sysctls: %s", err)

	By("Removing old test case results")

	err = os.RemoveAll(n.BasePath)
	Expect(err).To(Succeed(), "Failed to remove old test case result directory: %s", err)

	By("Creating directory for new test case results")

	err = os.MkdirAll(n.BasePath, 0o755)
	Expect(err).To(Succeed(), "Failed to create test case result directory: %s", err)

	// Ginkgo log
	logger, err = test.SetupLoggingWithFile(logFilename, true)
	Expect(err).To(Succeed())

	n.AgentOptions = append(n.AgentOptions,
		gopt.RedirectToLog(false),
	)

	n.NetworkOptions = append(n.NetworkOptions,
		gopt.Persistent(options.persist),
	)

	if options.capture {
		n.NetworkOptions = append(n.NetworkOptions,
			g.NewCapture(
				copt.Filename(pcapFilename),
				copt.LogKeys(true),
			),
		)
	}
}
