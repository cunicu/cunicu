// Package hsync synchronizes /etc/hosts with pairs of peer hostname and their respective IP addresses
package hsync

import (
	"fmt"
	"net/netip"
	"strings"

	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/util"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

const (
	hostsCommentPrefix = "cunicu"
	hostsPath          = "/etc/hosts"
)

func init() {
	daemon.Features["hsync"] = &daemon.FeaturePlugin{
		New:         New,
		Description: "Hosts synchronization",
		Order:       200,
	}
}

type Interface struct {
	*daemon.Interface

	logger *zap.Logger
}

func New(i *daemon.Interface) (daemon.Feature, error) {
	if !i.Settings.SyncHosts {
		return nil, nil
	}

	hs := &Interface{
		Interface: i,
		logger:    zap.L().Named("hsync").With(zap.String("intf", i.Name())),
	}

	i.OnPeer(hs)

	return hs, nil
}

func (hs *Interface) Start() error {
	hs.logger.Info("Started /etc/hosts synchronization")

	return nil
}

func (hs *Interface) Close() error {
	return hs.Update(nil)
}

func (hs *Interface) Hosts() []Host {

	d := hs.Settings.Domain
	if d != "" && !strings.HasPrefix(d, ".") {
		d = "." + d
	}

	hosts := []Host{}

	for _, p := range hs.Peers {
		m := map[netip.Addr][]string{}

		// Host names
		for name, addrs := range p.Hosts {
			for _, a := range addrs {
				// TODO: validate that the addresses are covered by the peers AllowedIPs
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
					p.Interface.KernelDevice.Name(),
					p.Interface.KernelDevice.Index(),
					p.PublicKey()),
			}

			hosts = append(hosts, h)
		}
	}

	return hosts
}

func (hs *Interface) Update(hosts []Host) error {
	lines, err := readLines(hostsPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Filter out lines not added by cunīcu
	lines = util.SliceFilter(lines, func(line string) bool {
		h, err := ParseHost(line)
		return err != nil || !strings.HasPrefix(h.Comment, hostsCommentPrefix) || !strings.Contains(h.Comment, fmt.Sprintf("ifindex=%d", hs.KernelDevice.Index()))
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

	hs.logger.Info("Updated hosts file", zap.Int("num_hosts", len(hosts)))

	return nil
}

func (hs *Interface) Sync() error {
	hosts := hs.Hosts()

	return hs.Update(hosts)
}
