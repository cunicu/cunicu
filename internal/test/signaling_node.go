//go:build linux
// +build linux

package test

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	g "github.com/stv0g/gont/pkg"
	"go.uber.org/zap"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/socket"
)

type SignalingNode struct {
	*g.Host

	Command *exec.Cmd
	Client  *socket.Client

	ID              peer.ID
	ListenAddresses []multiaddr.Multiaddr
}

func NewSignalingNode(m *g.Network, name string, opts ...g.Option) (*SignalingNode, error) {
	h, err := m.AddHost(name, opts...)
	if err != nil {
		return nil, err
	}

	b := &SignalingNode{
		Host:            h,
		ListenAddresses: []multiaddr.Multiaddr{},
	}

	return b, nil
}

func (s *SignalingNode) Start() error {
	var err error
	var stdout, stderr io.Reader
	var sockPath = fmt.Sprintf("/var/run/wice.%s.sock", s.Name())
	var logPath = fmt.Sprintf("logs/%s.log", s.Name())

	if err := os.RemoveAll(sockPath); err != nil {
		return fmt.Errorf("failed to remove old control socket: %w", err)
	}

	if stdout, stderr, s.Command, err = s.StartGo("../cmd/wice",
		"-socket", sockPath,
		"-socket-wait",
	); err != nil {
		return err
	}

	if _, err := FileWriter(logPath, stdout, stderr); err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	if s.Client, err = socket.Connect(sockPath); err != nil {
		return fmt.Errorf("failed to open control socket connection: %w", err)
	}

	if err := s.WaitReady(); err != nil {
		return err
	}

	return nil
}

func (s *SignalingNode) Stop() error {
	if s.Command == nil || s.Command.Process == nil {
		return nil
	}

	return s.Command.Process.Kill()
}

func (s *SignalingNode) Close() error {
	if err := s.Client.Close(); err != nil {
		return fmt.Errorf("failed to close RPC connection: %s", err)
	}

	return s.Stop()
}

func (s *SignalingNode) URL() (*url.URL, error) {
	pi := &peer.AddrInfo{
		ID:    s.ID,
		Addrs: s.ListenAddresses,
	}

	mas, err := peer.AddrInfoToP2pAddrs(pi)
	if err != nil {
		return nil, fmt.Errorf("failed to get p2p addresses")
	}

	q := url.Values{}
	// q.Add("dht", "false")
	// q.Add("mdns", "false")

	for _, ma := range mas {
		q.Add("bootstrap-peers", ma.String())
	}

	return &url.URL{
		Scheme:   "p2p",
		RawQuery: q.Encode(),
	}, nil
}

func (s *SignalingNode) WaitReady() error {
	var err error

	evt := s.Client.WaitForEvent(&pb.Event{
		Type:  "backend",
		State: "ready",
	})

	if be, ok := evt.Event.(*pb.Event_Backend); ok {
		s.ID, err = peer.Decode(be.Backend.Id)
		if err != nil {
			return fmt.Errorf("failed to decode peer ID: %w", err)
		}

		for _, la := range be.Backend.ListenAddresses {
			if ma, err := multiaddr.NewMultiaddr(la); err != nil {
				return fmt.Errorf("failed to decode listen address: %w", err)
			} else {
				s.ListenAddresses = append(s.ListenAddresses, ma)
			}
		}
	} else {
		zap.L().Warn("Missing signaling details")
	}

	return nil
}
