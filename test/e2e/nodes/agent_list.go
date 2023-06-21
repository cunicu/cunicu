// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package nodes

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/stv0g/cunicu/pkg/log"
	"github.com/stv0g/cunicu/test"
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
			if err := cb(a); err != nil {
				return fmt.Errorf("%w (node %s)", err, a.Name())
			}
			return nil
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

func (al AgentList) ForEachInterfacePairOneDir(cb func(a, b *WireGuardInterface) error) error {
	g := errgroup.Group{}

	for i, n := range al {
		for _, p := range al[i+1:] {
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

	return g.Wait()
}

func (al AgentList) WaitConnectionsReady(ctx context.Context) error {
	handler := &test.DefaultProgressHandler{
		Logger: log.Global.Named("wait-conns"),
	}

	return test.WithProgress(ctx, func(started, completed chan string) error {
		return al.ForEachInterfacePair(func(a, b *WireGuardInterface) error {
			id := fmt.Sprintf("%s <-> %s", a, b)

			started <- id
			err := a.WaitConnectionEstablished(ctx, b)
			completed <- id

			return err
		})
	}, handler)
}

func (al AgentList) PingPeers(ctx context.Context) error {
	handler := &test.DefaultProgressHandler{
		Logger: log.Global.Named("ping"),
	}

	return test.WithProgress(ctx, func(started, completed chan string) error {
		return al.ForEachInterfacePairOneDir(func(a, b *WireGuardInterface) error {
			id := fmt.Sprintf("%s <-> %s", a, b)

			started <- id
			err := a.PingPeer(ctx, b)
			completed <- id

			return err
		})
	}, handler)
}
