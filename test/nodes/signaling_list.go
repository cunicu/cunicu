package nodes

import (
	"fmt"

	g "github.com/stv0g/gont/pkg"
)

type SignalingNodeList []SignalingNode

func AddSignalingNodes(n *g.Network, numNodes int, opts ...g.Option) (SignalingNodeList, error) {
	nodes := SignalingNodeList{}

	for i := 1; i <= numNodes; i++ {
		node, err := NewGrpcSignalingNode(n, fmt.Sprintf("n%d", i))
		if err != nil {
			return nil, fmt.Errorf("failed to create relay: %w", err)
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (nl *SignalingNodeList) Start() error {
	for _, n := range *nl {
		if err := n.Start(); err != nil {
			return err
		}
	}

	return nil
}

func (nl *SignalingNodeList) Stop() error {
	for _, n := range *nl {
		if err := n.Stop(); err != nil {
			return err
		}
	}

	return nil
}
