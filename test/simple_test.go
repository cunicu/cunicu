//go:build linux
// +build linux

package main_test

import (
	"testing"

	g "github.com/stv0g/gont/pkg"
	gopt "github.com/stv0g/gont/pkg/options"
	"go.uber.org/zap"
	"riasc.eu/wice/internal/test"
)

func TestSimple(t *testing.T) {
	var (
		n  *g.Network
		sw *g.Switch
		s  *test.SignalingNode
		r  *test.RelayNode
		nl test.NodeList

		err error
	)

	logger := zap.L()

	if n, err = g.NewNetwork("", gopt.Persistent(true)); err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}
	defer n.Close()

	if sw, err = n.AddSwitch("sw1"); err != nil {
		t.Fatalf("Failed to create switch: %s", err)
	}

	if r, err = test.NewRelayNode(n, "r1"); err != nil {
		t.Fatalf("Failed to start relay: %s", err)
	}
	defer r.Close()

	if s, err = test.NewSignalingNode(n, "s1"); err != nil {
		t.Fatalf("Failed to create signaling node: %s", err)
	}
	defer s.Close()

	if nl, err = test.AddNodes(n, s, 2); err != nil {
		t.Fatalf("Failed to created nodes: %s", err)
	}

	n1 := nl[0]
	n2 := nl[1]

	if err := n.AddLink(
		gopt.Interface("eth0", gopt.AddressIPv4(10, 0, 0, 1, 24), r.Host),
		gopt.Interface("eth0-r", sw),
	); err != nil {
		t.Fatalf("Failed to add link: %s", err)
	}

	if err := n.AddLink(
		gopt.Interface("eth0", gopt.AddressIPv4(10, 0, 0, 2, 24), s.Host),
		gopt.Interface("eth0-s", sw),
	); err != nil {
		t.Fatalf("Failed to add link: %s", err)
	}

	if err := n.AddLink(
		gopt.Interface("eth0", gopt.AddressIPv4(10, 0, 0, 100, 24), n1.Host),
		gopt.Interface("eth0-n1", sw),
	); err != nil {
		t.Fatalf("Failed to add link: %s", err)
	}

	if err := n.AddLink(
		gopt.Interface("eth0", gopt.AddressIPv4(10, 0, 0, 101, 24), n2.Host),
		gopt.Interface("eth0-n2", sw),
	); err != nil {
		t.Fatalf("Failed to add link: %s", err)
	}

	if err := r.Start(); err != nil {
		t.Fatalf("Failed to start relay: %s", err)
	}

	if err := s.Start(); err != nil {
		t.Fatalf("Failed to start signaling node: %s", err)
	}

	args := []interface{}{
		// ICE options
		"-ice-user", r.Username,
		"-ice-pass", r.Password,

		// Limititing ourself to IPv4 network types
		"-ice-network-type", "tcp4",
		"-ice-network-type", "udp4",
	}
	for _, u := range r.URLs() {
		args = append(args, "-url", u)
	}

	if err := nl.StartAndWait(args...); err != nil {
		t.Fatalf("Failed to start WICE: %s", err)
	}
	defer nl.Stop()

	nl.ForEachPeer(func(n *test.Node) error {
		out, _, _ := n.Run("wg")
		logger.Info(string(out))

		out, _, _ = n.Run("ip", "a")
		logger.Info(string(out))

		return nil
	})

	if err := nl.PingPeers(); err != nil {
		t.Fatalf("Failed to ping peers: %s", err)
	}
}
