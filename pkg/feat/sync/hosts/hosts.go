// Package hosts synchronizes /etc/hosts with pairs of peer hostname and their respective link-local IP addresses
package hosts

import (
	"fmt"
	"net"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/util"
	"riasc.eu/wice/pkg/watcher"
)

const (
	hostsCommentPrefix = "wice"
	hostsPath          = "/etc/hosts"
)

type HostsSync struct {
	watcher *watcher.Watcher
	logger  *zap.Logger
}

func New(w *watcher.Watcher) (*HostsSync, error) {
	s := &HostsSync{
		watcher: w,
		logger:  zap.L().Named("sync.hosts"),
	}

	w.OnPeer(s)

	return s, nil
}

func (s *HostsSync) hosts() []Host {
	hosts := []Host{}

	for _, i := range s.watcher.Interfaces {
		for _, p := range i.Peers {
			if p.Name == "" {
				continue
			}

			// We use a shorted version of the public key as a DNS name here
			pkName := p.PublicKey().String()[:8]

			h := Host{
				IP:    p.PublicKey().IPv6Address().IP,
				Names: []string{pkName},
				Comment: fmt.Sprintf("%s: ifname=%s, ifindex=%d, pk=%s", hostsCommentPrefix,
					i.KernelDevice.Name(),
					i.KernelDevice.Index(),
					p.PublicKey().String()),
			}

			if p.Name != "" {
				h.Names = append(h.Names, p.Name)
			}

			hosts = append(hosts, h)
		}
	}

	return hosts
}

func (s *HostsSync) updateHostsFile() error {
	lines, err := readLines(hostsPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Filter out lines not added by wice
	lines = util.FilterSlice(lines, func(line string) bool {
		h, err := ParseHost(line)
		return err != nil || !strings.HasPrefix(h.Comment, hostsCommentPrefix)
	})

	// Add a separating new line
	lines = append(lines, "")

	// Add new hosts
	hosts := s.hosts()
	for _, h := range hosts {
		line, err := h.Line()
		if err != nil {
			return err
		}

		lines = append(lines, line)
	}

	// Remove double new lines
	lines = slices.Compact(lines)

	if err := writeLines(hostsPath, lines); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	s.logger.Info("Updated hosts file", zap.Int("num_hosts", len(hosts)))

	return nil
}

func (s *HostsSync) OnPeerAdded(p *core.Peer) {
	if err := s.updateHostsFile(); err != nil {
		s.logger.Error("Failed to update hosts file", zap.Error(err))
	}
}

func (s *HostsSync) OnPeerRemoved(p *core.Peer) {
	if err := s.updateHostsFile(); err != nil {
		s.logger.Error("Failed to update hosts file", zap.Error(err))
	}
}

func (s *HostsSync) OnPeerModified(p *core.Peer, old *wgtypes.Peer, m core.PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {
	// Only update if the name has changed
	if m.Is(core.PeerModifiedName) {
		if err := s.updateHostsFile(); err != nil {
			s.logger.Error("Failed to update hosts file", zap.Error(err))
		}
	}
}
