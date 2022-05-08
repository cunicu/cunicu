package net

import (
	g "github.com/stv0g/gont/pkg"
	"riasc.eu/wice/test/nodes"
)

type NetworkFactory func(p *NetworkParams) (*Network, error)

type NetworkParams struct {
	NetworkOptions []g.Option
	HostOptions    []g.Option

	NumAgents int
}

type Network struct {
	*g.Network

	Switch         *g.Switch
	SignalingNodes nodes.SignalingNodeList
	Relays         nodes.RelayList
	Agents         nodes.AgentList
}

func (n *Network) Close() error {
	if err := n.Agents.Stop(); err != nil {
		return err
	}

	if err := n.SignalingNodes.Stop(); err != nil {
		return err
	}

	if err := n.Relays.Stop(); err != nil {
		return err
	}

	return n.Network.Close()
}
