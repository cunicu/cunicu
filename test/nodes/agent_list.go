//go:build linux

package nodes

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

type AgentList []*Agent

func (al AgentList) Start(dir string, extraArgs ...any) error {
	if err := al.ForEachAgent(func(a *Agent) error {
		return a.Start("", dir, extraArgs...)
	}); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	return nil
}

func (al AgentList) Close() error {
	return al.ForEachAgent(func(a *Agent) error {
		return a.Close()
	})
}

func (al AgentList) ForEachAgent(cb func(a *Agent) error) error {
	g := errgroup.Group{}

	for _, a := range al {
		a := a

		g.Go(func() error {
			return cb(a)
		})
	}

	return g.Wait()
}

func (al AgentList) ForEachInterface(cb func(i *WireGuardInterface) error) error {
	g := errgroup.Group{}

	for _, n := range al {
		for _, ni := range n.WireGuardInterfaces {
			ni := ni // avoid aliasing

			g.Go(func() error {
				return cb(ni)
			})

		}
	}

	return g.Wait()
}

func (al AgentList) ForEachAgentPair(cb func(a, b *Agent) error) error {
	g := errgroup.Group{}

	for _, n := range al {
		for _, p := range al {
			if n != p {
				p := p // avoid aliasing
				n := n

				g.Go(func() error {
					return cb(n, p)
				})
			}
		}
	}

	return g.Wait()
}

func (al AgentList) ForEachInterfacePair(cb func(a, b *WireGuardInterface) error) error {
	g := errgroup.Group{}

	for _, n := range al {
		for _, p := range al {
			if n != p {
				for _, ni := range n.WireGuardInterfaces {
					for _, pi := range p.WireGuardInterfaces {
						pi := pi // avoid aliasing
						ni := ni

						g.Go(func() error {
							return cb(ni, pi)
						})
					}
				}
			}
		}
	}

	return g.Wait()
}

func (al AgentList) WaitConnectionsReady(ctx context.Context) error {
	return al.ForEachInterfacePair(func(a, b *WireGuardInterface) error {
		return a.WaitConnectionReady(ctx, b)
	})
}

func (al AgentList) PingPeers(ctx context.Context) error {
	return al.ForEachInterfacePair(func(a, b *WireGuardInterface) error {
		return a.PingPeer(ctx, b)
	})
}
