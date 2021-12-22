//go:build linux
// +build linux

package test

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	g "github.com/stv0g/gont/pkg"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/socket"
)

type SignalingNode struct {
	*g.Host

	logger logrus.FieldLogger

	Command *exec.Cmd

	ID              peer.ID
	ListenAddresses []multiaddr.Multiaddr

	Client *socket.Client
}

func NewSignalingNode(m *g.Network, name string, opts ...g.Option) (*SignalingNode, error) {
	h, err := m.AddHost(name, opts...)
	if err != nil {
		return nil, err
	}

	b := &SignalingNode{
		Host:            h,
		ListenAddresses: []multiaddr.Multiaddr{},
		logger:          logrus.WithField("node", h.Name),
	}

	// Prepare listen address
	ma, err := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/40151")
	if err != nil {
		return nil, err
	}

	b.ListenAddresses = append(b.ListenAddresses, ma)

	return b, nil
}

func (b *SignalingNode) Start() error {
	var err error
	var stdout, stderr io.Reader
	var sockPath = fmt.Sprintf("/var/run/wice.%s.sock", b.Name())
	var logPath = fmt.Sprintf("logs/%s.log", b.Name())

	if err := os.RemoveAll(sockPath); err != nil {
		log.Fatal(err)
	}

	if stdout, stderr, b.Command, err = b.StartGo("../cmd/wice",
		"-socket", sockPath,
		"-socket-wait",
	); err != nil {
		return err
	}

	if _, err := FileWriter(logPath, stdout, stderr); err != nil {
		return err
	}

	if b.Client, err = socket.Connect(sockPath); err != nil {
		return err
	}

	if err := b.WaitReady(); err != nil {
		return err
	}

	return nil
}

func (b *SignalingNode) Stop() error {
	if b.Command == nil || b.Command.Process == nil {
		return nil
	}

	return b.Command.Process.Kill()
}

func (b *SignalingNode) Close() error {
	if err := b.Client.Close(); err != nil {
		return fmt.Errorf("failed to close RPC connection: %s", err)
	}

	return b.Stop()
}

func (b *SignalingNode) URL() (*url.URL, error) {
	// pi := &peer.AddrInfo{
	// 	ID:    b.ID,
	// 	Addrs: b.ListenAddresses,
	// }

	// mas, err := peer.AddrInfoToP2pAddrs(pi)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get p2p addresses")
	// }

	q := url.Values{}
	// q.Add("dht", "false")
	// q.Add("mdns", "false")

	// for _, ma := range mas {
	// 	q.Add("bootstrap-peers", ma.String())
	// }

	return &url.URL{
		Scheme:   "p2p",
		RawQuery: q.Encode(),
	}, nil
}

func (b *SignalingNode) WaitReady() error {
	var err error

	evt := b.Client.WaitForEvent(&pb.Event{
		Type:  "backend",
		State: "ready",
	})

	be, ok := evt.Event.(*pb.Event_Backend)
	if ok {
		b.ID, err = peer.Decode(be.Backend.Id)
		if err != nil {
			return fmt.Errorf("failed to decode peer ID: %w", err)
		}

		for _, la := range be.Backend.ListenAddresses {
			if ma, err := multiaddr.NewMultiaddr(la); err != nil {
				return fmt.Errorf("failed to decode listen address: %w", err)
			} else {
				b.ListenAddresses = append(b.ListenAddresses, ma)
			}
		}
	} else {
		logrus.Warn("Missing signaling details")
	}

	return nil
}
