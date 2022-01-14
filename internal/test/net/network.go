package net

import (
	g "github.com/stv0g/gont/pkg"
	"riasc.eu/wice/internal/test"
)

type NetworkFactory func(p *NetworkParams) (*Network, error)

type NetworkParams struct {
	NetworkOptions []g.Option
	NumNodes       int
}

type Network struct {
	*g.Network

	Switch        *g.Switch
	SignalingNode *test.SignalingNode
	RelayNode     *test.RelayNode
	Nodes         test.NodeList
}

func (n *Network) Close() error {
	if err := n.Nodes.Stop(); err != nil {
		return err
	}

	if err := n.SignalingNode.Close(); err != nil {
		return err
	}

	if err := n.RelayNode.Close(); err != nil {
		return err
	}

	if err := n.Network.Close(); err != nil {
		return err
	}

	return nil
}
