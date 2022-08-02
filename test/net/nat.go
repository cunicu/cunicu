//go:build linux

package net

import (
	"fmt"
	"net"

	g "github.com/stv0g/gont/pkg"
	gopt "github.com/stv0g/gont/pkg/options"
	"riasc.eu/wice/test/nodes"
)

func NAT(p *NetworkParams) (*Network, error) {
	var (
		n  *g.Network
		sw *g.Switch
		r  nodes.RelayNode
		s  nodes.SignalingNode
		al nodes.AgentList

		err error
	)

	if n, err = g.NewNetwork("", p.NetworkOptions...); err != nil {
		return nil, fmt.Errorf("failed to create network: %w", err)
	}

	if sw, err = n.AddSwitch("sw1"); err != nil {
		return nil, fmt.Errorf("failed to create switch: %w", err)
	}

	if r, err = nodes.NewCoturnNode(n, "r1"); err != nil {
		return nil, fmt.Errorf("failed to start relay: %w", err)
	}

	if s, err = nodes.NewGrpcSignalingNode(n, "s1"); err != nil {
		return nil, fmt.Errorf("fFailed to create signaling node: %w", err)
	}

	if al, err = nodes.NewAgents(n, p.NumAgents); err != nil {
		return nil, fmt.Errorf("failed to created nodes: %w", err)
	}

	if err := n.AddLink(
		gopt.Interface("eth0", r,
			gopt.AddressIPv4(10, 0, 0, 1, 16),
			gopt.AddressIP("fc::1/64")),
		gopt.Interface("eth0-r", sw),
	); err != nil {
		return nil, fmt.Errorf("fFailed to add link: %w", err)
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
		sws, err := n.AddSwitch(fmt.Sprintf("sw1%d", i))
		if err != nil {
			return nil, err
		}

		_, err = n.AddNAT(fmt.Sprintf("nat%d", i),
			gopt.Interface("eth-nb", sw,
				gopt.NorthBound,
				gopt.AddressIPv4(10, 0, 1, byte(i)+1, 16),
				gopt.AddressIP(fmt.Sprintf("fc::1:%d/64", i))),
			gopt.Interface("eth-sb", sws,
				gopt.SouthBound,
				gopt.AddressIPv4(10, 1, 0, 1, 24),
				gopt.AddressIP(fmt.Sprintf("fc:1::%d/64", i))),
		)
		if err != nil {
			return nil, err
		}

		if err := n.AddLink(
			gopt.Interface("eth0", al[i].Host,
				gopt.AddressIPv4(10, 1, 0, 2, 24),
				gopt.AddressIP("fc:1::2/64")),
			gopt.Interface(fmt.Sprintf("eth0-n%d", i), sws),
		); err != nil {
			return nil, fmt.Errorf("failed to add link: %w", err)
		}

		if err := al[i].AddDefaultRoute(net.IPv4(10, 1, 0, 1)); err != nil {
			return nil, fmt.Errorf("failed to add route: %w", err)
		}

		if err := al[i].AddDefaultRoute(net.ParseIP("fc:1::1")); err != nil {
			return nil, fmt.Errorf("failed to add route: %w", err)
		}
	}

	return &Network{
		Network: n,
		Agents:  al,
		Switch:  sw,
	}, nil
}
