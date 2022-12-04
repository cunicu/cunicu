//nolint:dupl
package nodes

import (
	"fmt"

	g "github.com/stv0g/gont/pkg"
)

type RelayList []RelayNode

func AddRelayNodes(n *g.Network, numNodes int, opts ...g.Option) (RelayList, error) {
	ns := RelayList{}

	for i := 1; i <= numNodes; i++ {
		n, err := NewCoturnNode(n, fmt.Sprintf("n%d", i))
		if err != nil {
			return nil, fmt.Errorf("failed to create relay: %w", err)
		}

		ns = append(ns, n)
	}

	return ns, nil
}

func (l RelayList) Start(dir string, extraArgs ...any) error {
	for _, n := range l {
		if err := n.Start("", dir, extraArgs...); err != nil {
			return err
		}
	}

	return nil
}

func (l RelayList) Close() error {
	for _, n := range l {
		if err := n.Close(); err != nil {
			return err
		}
	}

	return nil
}
