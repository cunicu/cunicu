//go:build linux

package test_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	g "github.com/stv0g/gont/pkg"
	gopt "github.com/stv0g/gont/pkg/options"
	copt "github.com/stv0g/gont/pkg/options/capture"
	"go.uber.org/zap"

	"riasc.eu/wice/pkg/test"
	"riasc.eu/wice/pkg/util"
	"riasc.eu/wice/test/nodes"
)

var (
	logger *zap.Logger
)

type Network struct {
	*g.Network

	Name string

	NetworkOptions            []g.Option
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

	if setup {
		Skip("Aborting test as only network setup has been requested")
	}

	if len(n.Captures) > 0 && n.Captures[0].LogKeys {
		n.StartHandshakeTracer()
	}

	By("Writing network hosts file")

	hfn := filepath.Join(n.BasePath, "hosts")
	hf, err := os.OpenFile(hfn, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	Expect(err).To(Succeed(), "Failed to open hosts file: %s", err)

	err = n.Network.WriteHostsFile(hf)
	Expect(err).To(Succeed(), "Failed to write hosts file: %s", err)

	err = hf.Close()
	Expect(err).To(Succeed(), "Failed to close hosts file: %s", err)

	By("Saving network nodes file")

	nfn := filepath.Join(n.BasePath, "nodes")
	nf, err := os.OpenFile(nfn, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	Expect(err).To(Succeed(), "Failed to open nodes file: %s", err)

	n.AgentNodes.ForEachInterface(func(i *nodes.WireGuardInterface) error {
		_, err := fmt.Fprintf(nf, "%s %s %s\n", i.Agent.Name(), i.Name, i.PrivateKey.PublicKey())
		return err
	})

	err = nf.Close()
	Expect(err).To(Succeed(), "Failed to close nodes file: %s", err)

	By("Starting relay nodes")

	err = n.RelayNodes.Start(n.BasePath)
	Expect(err).To(Succeed(), "Failed to start relay: %s", err)

	By("Starting signaling nodes")

	err = n.SignalingNodes.Start(n.BasePath)
	Expect(err).To(Succeed(), "Failed to start signaling node: %s", err)

	By("Starting agent nodes")

	err = n.AgentNodes.Start(n.BasePath, n.AgentArgs()...)
	Expect(err).To(Succeed(), "Failed to start ɯice: %s", err)
}

func (n *Network) AgentArgs() []any {
	extraArgs := []any{}

	if len(n.RelayNodes) > 0 {
		// TODO: We currently assume that all relays use the same credentials
		extraArgs = append(extraArgs,
			"--username", n.RelayNodes[0].Username(),
			"--password", n.RelayNodes[0].Password(),
		)
	}

	for _, r := range n.RelayNodes {
		for _, u := range r.URLs() {
			extraArgs = append(extraArgs, "--url", u)
		}
	}

	for _, s := range n.SignalingNodes {
		extraArgs = append(extraArgs, "--backend", s.URL())
	}

	return extraArgs
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

	GinkgoWriter.ClearTeeWriters()
}

func (n *Network) ConnectivityTests() {
	It("", func() {
		By("Waiting until all peers are connected")

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		err := n.AgentNodes.WaitConnectionsReady(ctx)
		Expect(err).To(Succeed(), "Failed to wait for peers to connect: %s", err)

		By("Ping between all peers")

		err = n.AgentNodes.PingPeers(ctx)
		Expect(err).To(Succeed(), "Failed to ping peers: %s", err)
	})
}

func (n *Network) Init() {
	*n = Network{}

	n.Name = fmt.Sprintf("wice-%d", rand.Uint32())

	name := GinkgoT().Name()
	n.BasePath = filepath.Join(strings.Split(name, " ")...)
	n.BasePath = filepath.Join("logs", n.BasePath)

	logFilename := filepath.Join(n.BasePath, "test.log")
	pcapFilename := filepath.Join(n.BasePath, "capture.pcapng")

	By("Tweaking sysctls for large Gont networks")

	err := util.SetSysctlMap(map[string]any{
		"net.ipv4.neigh.default.gc_thresh1": 10000,
		"net.ipv4.neigh.default.gc_thresh2": 15000,
		"net.ipv4.neigh.default.gc_thresh3": 20000,
		"net.ipv6.neigh.default.gc_thresh1": 10000,
		"net.ipv6.neigh.default.gc_thresh2": 15000,
		"net.ipv6.neigh.default.gc_thresh3": 20000,
	})
	Expect(err).To(Succeed(), "Failed to set sysctls: %s", err)

	By("Removing old test case results")

	err = os.RemoveAll(n.BasePath)
	Expect(err).To(Succeed(), "Failed to remove old test case result directory: %s", err)

	By("Creating directory for new test case results")

	err = os.MkdirAll(n.BasePath, 0755)
	Expect(err).To(Succeed(), "Failed to create test case result directory: %s", err)

	// Ginkgo log
	logger = test.SetupLoggingWithFile(logFilename, true)

	n.AgentOptions = append(n.AgentOptions,
		gopt.LogToDebug(false),
	)

	n.NetworkOptions = append(n.NetworkOptions,
		gopt.Persistent(persist),
	)

	if capture {
		n.NetworkOptions = append(n.NetworkOptions,
			gopt.CaptureAll(
				copt.Filename(pcapFilename),
				copt.LogKeys(true),
			),
		)
	}
}
