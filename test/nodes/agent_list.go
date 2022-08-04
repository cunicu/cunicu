//go:build linux

package nodes

import (
	"fmt"

	"golang.org/x/sync/errgroup"
)

type AgentList []*Agent

func (al AgentList) Start(args []any) error {
	if err := al.ForEachAgentPair(func(a, b *Agent) error {
		return a.AddWireGuardPeer(b)
	}); err != nil {
		return fmt.Errorf("failed to add WireGuard peers: %w", err)
	}

	if err := al.ForEachAgent(func(a *Agent) error {
		return a.Start(args)
	}); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	return nil
}

func (al AgentList) Stop() error {
	return al.ForEachAgent(func(a *Agent) error {
		return a.Stop()
	})
}

func (al AgentList) ForEachAgent(cb func(a *Agent) error) error {
	g := errgroup.Group{}

	for _, node := range al {
		n := node

		g.Go(func() error {
			return cb(n)
		})
	}

	return g.Wait()
}

func (al AgentList) ForEachAgentPair(cb func(a, b *Agent) error) error {
	g := errgroup.Group{}

	for _, n := range al {
		for _, p := range al {
			if n != p {
				peer := p
				node := n

				g.Go(func() error {
					return cb(node, peer)
				})
			}
		}
	}

	return g.Wait()
}

func (al AgentList) WaitConnected() error {
	return al.ForEachAgentPair(func(a, b *Agent) error {
		return a.WaitReady(b)
	})
}

func (al AgentList) PingPeers() error {
	return al.ForEachAgentPair(func(a, b *Agent) error {
		return a.PingWireGuardPeer(b)
	})
}
