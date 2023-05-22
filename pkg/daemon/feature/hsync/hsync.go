// Package hsync synchronizes /etc/hosts with pairs of peer hostname and their respective IP addresses
package hsync

import (
	"fmt"
	"net/netip"
	"strings"

	"github.com/stv0g/cunicu/pkg/daemon"
	slicesx "github.com/stv0g/cunicu/pkg/types/slices"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

const (
	hostsCommentPrefix = "cunicu"
	hostsPath          = "/etc/hosts"
)

var Get = daemon.RegisterFeature(New, 200) //nolint:gochecknoglobals

type Interface struct {
	*daemon.Interface

	logger *zap.Logger
}

func New(i *daemon.Interface) (*Interface, error) {
	if !i.Settings.SyncHosts {
		return nil, daemon.ErrFeatureDeactivated
	}

	hs := &Interface{
		Interface: i,
		logger:    zap.L().Named("hsync").With(zap.String("intf", i.Name())),
	}

	i.AddPeerHandler(hs)

	return hs, nil
}

func (i *Interface) Start() error {
	i.logger.Info("Started /etc/hosts synchronization")

	return nil
}

func (i *Interface) Close() error {
	return i.Update(nil)
}

func (i *Interface) Hosts() []Host {
	d := i.Settings.Domain
	if d != "" && !strings.HasPrefix(d, ".") {
		d = "." + d
	}

	hosts := []Host{}

	for _, p := range i.Peers {
		m := map[netip.Addr][]string{}

		for name, addrs := range p.Hosts {
			for _, a := range addrs {
				// TODO: Validate that the addresses are covered by the peers AllowedIPs
				addr, ok := netip.AddrFromSlice(a)
				if !ok {
					continue
				}

				m[addr] = append(m[addr], name+d)
			}
		}

		for addr, names := range m {
			h := Host{
				Names: names,
				IP:    addr.AsSlice(),
				Comment: fmt.Sprintf("%s: ifname=%s, ifindex=%d, pk=%s", hostsCommentPrefix,
					p.Interface.Name(),
					p.Interface.Index(),
					p.PublicKey()),
			}

			hosts = append(hosts, h)
		}
	}

	return hosts
}

func (i *Interface) Update(hosts []Host) error {
	lines, err := readLines(hostsPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Filter out lines not added by cunīcu
	lines = slicesx.Filter(lines, func(line string) bool {
		h, err := ParseHost(line)
		return err != nil || !strings.HasPrefix(h.Comment, hostsCommentPrefix) || !strings.Contains(h.Comment, fmt.Sprintf("ifindex=%d", i.Index()))
	})

	// Add new hosts
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

	i.logger.Info("Updated hosts file", zap.Int("num_hosts", len(hosts)))

	return nil
}

func (i *Interface) Sync() error {
	hosts := i.Hosts()

	return i.Update(hosts)
}
