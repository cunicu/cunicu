// Package hsync synchronizes /etc/hosts with pairs of peer hostname and their respective link-local IP addresses
package hsync

import (
	"fmt"
	"net"
	"strings"

	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/util"
	"github.com/stv0g/cunicu/pkg/watcher"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	hostsCommentPrefix = "cunicu"
	hostsPath          = "/etc/hosts"
)

type HostsSync struct {
	domain string

	watcher *watcher.Watcher
	logger  *zap.Logger
}

func New(w *watcher.Watcher, d string) *HostsSync {
	if d != "" && !strings.HasPrefix(d, ".") {
		d = "." + d
	}

	hs := &HostsSync{
		domain:  d,
		watcher: w,
		logger:  zap.L().Named("hsync"),
	}

	w.OnPeer(hs)

	return hs
}

func (hs *HostsSync) Start() error {
	hs.logger.Info("Started /etc/hosts synchronization")

	return nil
}

func (hs *HostsSync) Close() error {
	return nil
}

func (hs *HostsSync) Hosts() []Host {
	hosts := []Host{}

	hs.watcher.ForEachPeer(func(p *core.Peer) error {
		// We use a shorted version of the public key as a DNS name here
		pkName := p.PublicKey().String()[:8]

		h := Host{
			IP: p.PublicKey().IPv6Address().IP,
			Names: []string{
				pkName + hs.domain,
			},
			Comment: fmt.Sprintf("%s: ifname=%s, ifindex=%d, pk=%s", hostsCommentPrefix,
				p.Interface.KernelDevice.Name(),
				p.Interface.KernelDevice.Index(),
				p.PublicKey()),
		}

		if p.Name != "" {
			h.Names = append(h.Names, p.Name+hs.domain)
		}

		hosts = append(hosts, h)

		return nil
	})

	return hosts
}

func (hs *HostsSync) Sync() error {
	lines, err := readLines(hostsPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Filter out lines not added by cunīcu
	lines = util.FilterSlice(lines, func(line string) bool {
		h, err := ParseHost(line)
		return err != nil || !strings.HasPrefix(h.Comment, hostsCommentPrefix)
	})

	// Add a separating new line
	lines = append(lines, "")

	// Add new hosts
	hosts := hs.Hosts()
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

	hs.logger.Info("Updated hosts file", zap.Int("num_hosts", len(hosts)))

	return nil
}

func (hs *HostsSync) OnPeerAdded(p *core.Peer) {
	if err := hs.Sync(); err != nil {
		hs.logger.Error("Failed to update hosts file", zap.Error(err))
	}
}

func (hs *HostsSync) OnPeerRemoved(p *core.Peer) {
	if err := hs.Sync(); err != nil {
		hs.logger.Error("Failed to update hosts file", zap.Error(err))
	}
}

func (hs *HostsSync) OnPeerModified(p *core.Peer, old *wgtypes.Peer, m core.PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {
	// Only update if the name has changed
	if m.Is(core.PeerModifiedName) {
		if err := hs.Sync(); err != nil {
			hs.logger.Error("Failed to update hosts file", zap.Error(err))
		}
	}
}
