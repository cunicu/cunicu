//go:build linux

package main_test

import (
	"os"
	"testing"

	"riasc.eu/wice/internal/test"
	"riasc.eu/wice/internal/test/net"

	g "github.com/stv0g/gont/pkg"
	gopt "github.com/stv0g/gont/pkg/options"
)

func TestSimpleUser(t *testing.T) {
	RunTest(t,
		net.Simple,
		&net.NetworkParams{
			NumNodes: 2,
		},
		[]interface{}{
			"--proxy", "user",
		},
	)
}

func TestSimpleIPv6(t *testing.T) {
	RunTest(t,
		net.Simple,
		&net.NetworkParams{
			NumNodes: 3,
		},
		[]interface{}{
			"--proxy", "user",
			"--ice-network-type", "tcp6",
			"--ice-network-type", "udp6",
		},
	)
}

func TestSimpleTURN(t *testing.T) {
	RunTest(t,
		net.Simple,
		&net.NetworkParams{
			NumNodes: 3,
		},
		[]interface{}{
			"--proxy", "user",
			"--ice-candidate-type", "relay",
		},
	)
}

func TestSimple(t *testing.T) {
	RunTest(t,
		net.Simple,
		&net.NetworkParams{
			NumNodes: 3,
		},
		[]interface{}{
			// Limititing ourself to IPv4 network types
			// "--ice-network-type", "tcp4",
			// "--ice-network-type", "udp4",

			// "--ice-candidate-type", "relay",
			// "--proxy", "user",
		},
	)
}

func TestNAT(t *testing.T) {
	RunTest(t,
		net.NAT,
		&net.NetworkParams{
			NetworkOptions: []g.Option{
				gopt.Persistent(true),
			},
			NumNodes: 2,
		},
		[]interface{}{},
	)
}

func RunTest(t *testing.T, factory net.NetworkFactory, p *net.NetworkParams, args []interface{}) {
	n, err := factory(p)
	if err != nil {
		t.Fatalf("Failed to setup network: %s", err)
	}
	defer n.Close()

	if err := n.RelayNode.Start(); err != nil {
		t.Fatalf("Failed to start relay: %s", err)
	}

	if err := n.SignalingNode.Start(); err != nil {
		t.Fatalf("Failed to start signaling node: %s", err)
	}

	args = append(args,
		"--ice-user", n.RelayNode.Username,
		"--ice-pass", n.RelayNode.Password,
	)
	for _, u := range n.RelayNode.URLs() {
		args = append(args, "--url", u)
	}

	if err := n.Nodes.Start(args...); err != nil {
		t.Fatalf("Failed to start WICE: %s", err)
	}
	defer n.Nodes.Stop()

	if err := n.Nodes.WaitConnected(); err != nil {
		t.Fatalf("Failed to wait for peers to connect: %s", err)
	}

	if err := n.Nodes.PingPeers(); err != nil {
		t.Errorf("Failed to ping peers: %s", err)
	}

	n.Nodes.ForEachPeer(func(n *test.Node) error {
		out, _, _ := n.Run("wg", "show")
		os.Stdout.Write(out)

		out, _, _ = n.Run("ip", "addr", "show")
		os.Stdout.Write(out)

		return nil
	})
}
