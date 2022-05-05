package net

import (
	"fmt"

	g "github.com/stv0g/gont/pkg"
	gopt "github.com/stv0g/gont/pkg/options"
	"riasc.eu/wice/test/e2e"
)

func Simple(p *NetworkParams) (*Network, error) {
	var (
		n  *g.Network
		sw *g.Switch
		s  e2e.SignalingNode
		r  e2e.RelayNode
		al e2e.AgentList

		err error
	)

	if n, err = g.NewNetwork("", p.NetworkOptions...); err != nil {
		return nil, fmt.Errorf("failed to create network: %w", err)
	}

	if sw, err = n.AddSwitch("sw1"); err != nil {
		return nil, fmt.Errorf("failed to create switch: %w", err)
	}

	if r, err = e2e.NewCoturnNode(n, "r1"); err != nil {
		return nil, fmt.Errorf("failed to start relay: %w", err)
	}

	if s, err = e2e.NewGrpcSignalingNode(n, "s1"); err != nil {
		return nil, fmt.Errorf("fFailed to create signaling node: %w", err)
	}

	if al, err = e2e.NewAgents(n, p.NumAgents, p.HostOptions...); err != nil {
		return nil, fmt.Errorf("failed to created nodes: %w", err)
	}

	if err := n.AddLink(
		gopt.Interface("eth0", r,
			gopt.AddressIPv4(10, 0, 0, 1, 16),
			gopt.AddressIP("fc::1/64")),
		gopt.Interface("eth0-r", sw),
	); err != nil {
		return nil, fmt.Errorf("failed to add link: %w", err)
	}

	if err := n.AddLink(
		gopt.Interface("eth0", s,
			gopt.AddressIPv4(10, 0, 0, 2, 16),
			gopt.AddressIP("fc::2/64")),
		gopt.Interface("eth0-s", sw),
	); err != nil {
		return nil, fmt.Errorf("failed to add link: %w", err)
	}

	for i := 0; i < p.NumAgents; i++ {
		if err := n.AddLink(
			gopt.Interface("eth0", al[i].Host,
				gopt.AddressIPv4(10, 0, 1, byte(i), 16),
				gopt.AddressIP(fmt.Sprintf("fc::1:%d/64", i))),
			gopt.Interface(fmt.Sprintf("eth0-n%d", i), sw),
		); err != nil {
			return nil, fmt.Errorf("failed to add link: %w", err)
		}
	}

	return &Network{
		Network:        n,
		Agents:         al,
		Relays:         e2e.RelayList{},
		SignalingNodes: e2e.SignalingNodeList{},
		Switch:         sw,
	}, nil
}
