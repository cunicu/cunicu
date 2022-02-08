//go:build linux

package e2e_test

import (
	"testing"

	"go.uber.org/zap"
	"riasc.eu/wice/internal/test"
	"riasc.eu/wice/test/e2e"
	"riasc.eu/wice/test/e2e/net"
)

func TestMain(m *testing.M) {
	test.Main(m)
}

func RunTest(t *testing.T, factory net.NetworkFactory, p *net.NetworkParams, args []interface{}) {
	logger := zap.L().Named("test.e2e")

	n, err := factory(p)
	if err != nil {
		t.Fatalf("Failed to setup network: %s", err)
	}
	defer n.Close()

	if err := n.Relays.Start(); err != nil {
		t.Fatalf("Failed to start relay: %s", err)
	}

	logger.Info("Starting signaling nodes", zap.Int("count", len(n.SignalingNodes)))
	if err := n.SignalingNodes.Start(); err != nil {
		t.Fatalf("Failed to start signaling node: %s", err)
	}

	logger.Info("Starting relay nodes", zap.Int("count", len(n.Relays)))
	if len(n.Relays) > 0 {
		args = append(args,
			"--ice-user", n.Relays[0].Username(),
			"--ice-pass", n.Relays[0].Password(),
		)
		for _, u := range n.Relays.URLs() {
			args = append(args, "--url", u)
		}
	}

	logger.Info("Starting agent nodes", zap.Int("count", len(n.Agents)))
	if err := n.Agents.Start(args...); err != nil {
		t.Fatalf("Failed to start WICE: %s", err)
	}
	defer n.Agents.Stop()

	logger.Info("Wait until connections are established")
	if err := n.Agents.WaitConnected(); err != nil {
		t.Fatalf("Failed to wait for peers to connect: %s", err)
	}

	logger.Info("Ping between peers")
	if err := n.Agents.PingPeers(); err != nil {
		t.Errorf("Failed to ping peers: %s", err)
	}

	n.Agents.ForEachAgent(func(a *e2e.Agent) error {
		a.Dump()
		return nil
	})
}
