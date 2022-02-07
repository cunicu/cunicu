//go:build linux

package e2e_test

import (
	"testing"

	"riasc.eu/wice/internal/test"
	"riasc.eu/wice/test/e2e"
	"riasc.eu/wice/test/e2e/net"
)

func TestMain(m *testing.M) {
	test.Main(m)
}

func RunTest(t *testing.T, factory net.NetworkFactory, p *net.NetworkParams, args []interface{}) {
	n, err := factory(p)
	if err != nil {
		t.Fatalf("Failed to setup network: %s", err)
	}
	defer n.Close()

	if err := n.Relays.Start(); err != nil {
		t.Fatalf("Failed to start relay: %s", err)
	}

	if err := n.SignalingNodes.Start(); err != nil {
		t.Fatalf("Failed to start signaling node: %s", err)
	}

	if len(n.Relays) > 0 {
		args = append(args,
			"--ice-user", n.Relays[0].Username(),
			"--ice-pass", n.Relays[0].Password(),
		)
		for _, u := range n.Relays.URLs() {
			args = append(args, "--url", u)
		}
	}

	t.Logf("Starting %d agents\n", len(n.Agents))
	if err := n.Agents.Start(args...); err != nil {
		t.Fatalf("Failed to start WICE: %s", err)
	}
	defer n.Agents.Stop()

	t.Logf("Wait until connections are established\n")
	if err := n.Agents.WaitConnected(); err != nil {
		t.Fatalf("Failed to wait for peers to connect: %s", err)
	}

	t.Logf("Ping between peers\n")
	if err := n.Agents.PingPeers(); err != nil {
		t.Errorf("Failed to ping peers: %s", err)
	}

	n.Agents.ForEachAgent(func(a *e2e.Agent) error {
		t.Logf("Details for agent %s\n", a.Name())

		a.DumpWireguardInterfaces()
		a.Run("ip", "addr", "show")

		return nil
	})
}
