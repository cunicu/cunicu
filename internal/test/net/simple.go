package net

import (
	"fmt"

	g "github.com/stv0g/gont/pkg"
	gopt "github.com/stv0g/gont/pkg/options"
	"riasc.eu/wice/internal/test"
)

func Simple(p *NetworkParams) (*Network, error) {
	var (
		n  *g.Network
		sw *g.Switch
		s  *test.SignalingNode
		r  *test.RelayNode
		nl test.NodeList

		err error
	)

	if n, err = g.NewNetwork("", gopt.Persistent(true)); err != nil {
		return nil, fmt.Errorf("failed to create network: %w", err)
	}
	defer n.Close()

	if sw, err = n.AddSwitch("sw1"); err != nil {
		return nil, fmt.Errorf("failed to create switch: %w", err)
	}

	if r, err = test.NewRelayNode(n, "r1"); err != nil {
		return nil, fmt.Errorf("failed to start relay: %w", err)
	}
	defer r.Close()

	if s, err = test.NewSignalingNode(n, "s1"); err != nil {
		return nil, fmt.Errorf("fFailed to create signaling node: %w", err)
	}
	defer s.Close()

	if nl, err = test.AddNodes(n, s, p.NumNodes); err != nil {
		return nil, fmt.Errorf("failed to created nodes: %w", err)
	}

	if err := n.AddLink(
		gopt.Interface("eth0", r.Host,
			gopt.AddressIPv4(10, 0, 0, 1, 16),
			gopt.AddressIP("fc::1/64")),
		gopt.Interface("eth0-r", sw),
	); err != nil {
		return nil, fmt.Errorf("fFailed to add link: %w", err)
	}

	if err := n.AddLink(
		gopt.Interface("eth0", s.Host,
			gopt.AddressIPv4(10, 0, 0, 2, 16),
			gopt.AddressIP("fc::2/64")),
		gopt.Interface("eth0-s", sw),
	); err != nil {
		return nil, fmt.Errorf("failed to add link: %w", err)
	}

	for i := 0; i < p.NumNodes; i++ {
		if err := n.AddLink(
			gopt.Interface("eth0", nl[i].Host,
				gopt.AddressIPv4(10, 0, 1, byte(i), 16),
				gopt.AddressIP(fmt.Sprintf("fc::1:%d/64", i))),
			gopt.Interface(fmt.Sprintf("eth0-n%d", i), sw),
		); err != nil {
			return nil, fmt.Errorf("failed to add link: %w", err)
		}
	}

	return &Network{
		Network:       n,
		Nodes:         nl,
		RelayNode:     r,
		SignalingNode: s,
		Switch:        sw,
	}, nil
}
