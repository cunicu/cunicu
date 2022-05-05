package e2e

import (
	"fmt"
	"net/url"

	"github.com/libp2p/go-libp2p-core/peer"
	"golang.org/x/sync/errgroup"
)

type AgentList []*Agent

func (al AgentList) Start(args ...interface{}) error {
	if err := al.ForEachAgentPair(func(a, b *Agent) error {
		return a.AddWireguardPeer(b)
	}); err != nil {
		return fmt.Errorf("failed to add wireguard peers")
	}

	if err := al.ForEachAgent(func(a *Agent) error {
		return a.Start(args...)
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
		return a.PingWireguardPeer(b)
	})
}

func (al AgentList) SignalingURL() (*url.URL, error) {
	q := url.Values{}
	// q.Add("dht", "false")
	q.Add("mdns", "false")

	for _, node := range al {
		pi := &peer.AddrInfo{
			ID:    node.ID,
			Addrs: node.ListenAddresses,
		}

		mas, err := peer.AddrInfoToP2pAddrs(pi)
		if err != nil {
			return nil, fmt.Errorf("failed to get p2p addresses")
		}

		for _, ma := range mas {
			q.Add("bootstrap-peer", ma.String())
		}
	}

	return &url.URL{
		Scheme:   "p2p",
		RawQuery: q.Encode(),
	}, nil
}
