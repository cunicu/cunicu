//go:build linux

package cases_test

import (
	g "github.com/stv0g/gont/pkg"
	options "github.com/stv0g/gont/pkg/options"
	"go.uber.org/zap"
	"riasc.eu/wice/test/net"
	"riasc.eu/wice/test/nodes"

	. "github.com/onsi/gomega"
)

var (
	GlobalNetworkOptions = []g.Option{
		options.Persistent(true),
	}
)

func Run(factory net.NetworkFactory, a *nodes.AgentParams, p *net.NetworkParams) {
	logger := zap.L().Named("test.e2e")

	if a.Arguments == nil {
		a.Arguments = []any{}
	}

	if p.NetworkOptions == nil {
		p.NetworkOptions = []g.Option{}
	}
	p.NetworkOptions = append(p.NetworkOptions, GlobalNetworkOptions...)

	n, err := factory(p)
	Expect(err).To(Succeed(), "Failed to create network: %s", err)
	defer n.Close()

	logger.Info("Starting relay nodes", zap.Int("count", len(n.Relays)))
	err = n.Relays.Start()
	Expect(err).To(Succeed(), "Failed to start relay: %s", err)

	logger.Info("Starting signaling nodes", zap.Int("count", len(n.SignalingNodes)))
	err = n.SignalingNodes.Start()
	Expect(err).To(Succeed(), "Failed to start signaling node: %s", err)

	if len(n.Relays) > 0 {
		// TODO: We currently assume that all relays use the same credentials
		a.Arguments = append(a.Arguments, "--username", n.Relays[0].Username())
		a.Arguments = append(a.Arguments, "--password", n.Relays[0].Password())
	}

	for _, r := range n.Relays {
		for _, u := range r.URLs() {
			a.Arguments = append(a.Arguments, "--url", u)
		}
	}

	for _, s := range n.SignalingNodes {
		a.Arguments = append(a.Arguments, "--backend", s.URL())
	}

	logger.Info("Starting agent nodes", zap.Int("count", len(n.Agents)))
	err = n.Agents.Start(a.Arguments)
	Expect(err).To(Succeed(), "Failed to start ɯice: %s", err)
	defer n.Agents.Stop()

	logger.Info("Wait until connections are established")
	err = n.Agents.WaitConnected()
	Expect(err).To(Succeed(), "Failed to wait for peers to connect: %s", err)

	logger.Info("Ping between peers")
	err = n.Agents.PingPeers()
	Expect(err).To(Succeed(), "Failed to ping peers: %s", err)
}
