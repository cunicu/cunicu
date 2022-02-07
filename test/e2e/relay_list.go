package e2e

import (
	"fmt"

	"github.com/pion/ice/v2"
	g "github.com/stv0g/gont/pkg"
)

type RelayList []RelayNode

func AddRelayNodes(n *g.Network, numNodes int, opts ...g.Option) (RelayList, error) {
	nodes := RelayList{}

	for i := 1; i <= numNodes; i++ {
		node, err := NewCoturnNode(n, fmt.Sprintf("n%d", i))
		if err != nil {
			return nil, fmt.Errorf("failed to create relay: %w", err)
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (nl *RelayList) Start() error {
	for _, n := range *nl {
		if err := n.Start(); err != nil {
			return err
		}
	}

	return nil
}

func (nl *RelayList) Stop() error {
	for _, n := range *nl {
		if err := n.Stop(); err != nil {
			return err
		}
	}

	return nil
}

func (nl RelayList) URLs() []*ice.URL {
	u := []*ice.URL{}

	for _, r := range nl {
		u = append(u, r.URLs()...)
	}

	return u
}
