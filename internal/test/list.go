//go:build linux

package test

import (
	"fmt"
	"net"

	g "github.com/stv0g/gont/pkg"
	"golang.org/x/sync/errgroup"
)

type NodeList []*Node

func AddNodes(n *g.Network, backend *SignalingNode, numNodes int, opts ...g.Option) (NodeList, error) {
	nodes := []*Node{}

	for i := 1; i <= numNodes; i++ {
		addr := net.IPNet{
			IP:   net.IPv4(172, 16, 0, byte(i)),
			Mask: net.IPv4Mask(255, 255, 0, 0),
		}

		node, err := NewNode(n, fmt.Sprintf("n%d", i), backend, addr, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create node: %w", err)
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

func ParseIP(s string) (net.IPNet, error) {
	ip, netw, err := net.ParseCIDR(s)
	if err != nil {
		return net.IPNet{}, err
	}

	return net.IPNet{
		IP:   ip,
		Mask: netw.Mask,
	}, nil
}

func (nl NodeList) Start(args ...interface{}) error {
	if err := nl.AddPeers(); err != nil {
		return err
	}

	if err := nl.ForEachPeer(func(n *Node) error {
		return n.Start(args...)
	}); err != nil {
		return err
	}

	return nil
}

func (nl NodeList) Stop() error {
	return nl.ForEachPeer(func(n *Node) error {
		return n.Stop()
	})
}

func (nl NodeList) ForEachPeer(cb func(n *Node) error) error {
	g := errgroup.Group{}

	for _, node := range nl {
		n := node

		g.Go(func() error {
			return cb(n)
		})
	}

	return g.Wait()
}

func (nl NodeList) ForEachPeerPair(cb func(a, b *Node) error) error {
	g := errgroup.Group{}

	for _, node := range nl {
		for _, peer := range nl {
			if peer == node {
				continue
			}

			n := node
			p := peer

			g.Go(func() error {
				return cb(n, p)
			})
		}
	}

	return g.Wait()
}

func (nl NodeList) WaitConnected() error {
	return nl.ForEachPeerPair(func(a, b *Node) error {
		return a.WaitReady(b)
	})
}

func (nl NodeList) AddPeers() error {
	return nl.ForEachPeerPair(func(a, b *Node) error {
		return a.AddPeer(b)
	})
}

func (nl NodeList) PingPeers() error {
	return nl.ForEachPeerPair(func(a, b *Node) error {
		return a.PingPeer(b)
	})
}
