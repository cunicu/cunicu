// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/link"
	"github.com/stv0g/cunicu/pkg/log"
	"github.com/stv0g/cunicu/pkg/wg"
)

var errNotSupported = errors.New("not supported on this platform")

type WireGuardProvider struct {
	path   string
	order  []string
	logger *log.Logger
}

func NewWireGuardProvider() *WireGuardProvider {
	path := os.Getenv("WG_CONFIG_PATH")
	if path == "" {
		path = wg.ConfigPath
	}

	return &WireGuardProvider{
		path: path,

		logger: log.Global.Named("config.wg"),
	}
}

func (p *WireGuardProvider) Read() (map[string]interface{}, error) {
	m := map[string]any{}

	des, err := os.ReadDir(p.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return m, nil
		}

		return nil, fmt.Errorf("failed to list config files in '%s': %w", p.path, err)
	}

	p.order = []string{}

	for _, de := range des {
		cfg := filepath.Join(p.path, de.Name())
		filename := filepath.Base(cfg)
		extension := filepath.Ext(filename)
		name := strings.TrimSuffix(filename, extension)

		if extension != ".conf" {
			continue
		}

		cfgData, err := os.ReadFile(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to open config file %s: %w", cfg, err)
		}

		var pa wgParser

		m[name], err = pa.Unmarshal(cfgData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse config file %s: %w", cfg, err)
		}

		p.order = append(p.order, name)
	}

	return map[string]any{
		"interfaces": m,
	}, nil
}

func (p *WireGuardProvider) ReadBytes() ([]byte, error) {
	return nil, errNotImplemented
}

func (p *WireGuardProvider) Order() []string {
	slices.Sort(p.order)

	return p.order
}

type wgParser struct{}

func (p *wgParser) Unmarshal(data []byte) (map[string]interface{}, error) {
	c, err := wg.ParseConfig(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	s, err := NewInterfaceSettingsFromConfig(c)
	if err != nil {
		return nil, err
	}

	return Map(s, "koanf"), nil
}

func (p *wgParser) Marshal(map[string]interface{}) ([]byte, error) {
	return nil, errNotSupported
}

func NewInterfaceSettingsFromConfig(c *wg.Config) (*InterfaceSettings, error) {
	s := &InterfaceSettings{
		ListenPort: c.ListenPort,
		Peers:      map[string]PeerSettings{},
		Addresses:  c.Address,
		DNS:        c.DNS,
	}

	if c.PrivateKey != nil {
		s.PrivateKey = crypto.Key(*c.PrivateKey)
	}

	if c.FirewallMark != nil {
		s.FirewallMark = *c.FirewallMark
	}

	if c.MTU != nil {
		s.MTU = *c.MTU
	}

	if c.Table != nil && *c.Table != "off" {
		var err error

		s.RoutingTable, err = link.Table(*c.Table)
		if err != nil {
			return nil, fmt.Errorf("failed to parse routing table '%s': %w", *c.Table, err)
		}
	}

	// TODO: Add exec hooks for wg-quick PreUp, PostUp, PreDown, PostDown hooks

	for i, p := range c.Peers {
		wgps := PeerSettings{
			PublicKey:  crypto.Key(p.PublicKey),
			AllowedIPs: p.AllowedIPs,
		}

		if p.PresharedKey != nil {
			wgps.PresharedKey = crypto.Key(*p.PresharedKey)
		}

		if c.PeerEndpoints[i] != "" {
			wgps.Endpoint = c.PeerEndpoints[i]
		}

		if p.PersistentKeepaliveInterval != nil {
			wgps.PersistentKeepaliveInterval = *p.PersistentKeepaliveInterval
		}

		name := c.PeerNames[i]
		if name == "" {
			name = p.PublicKey.String()[:8]
		}

		s.Peers[name] = wgps
	}

	return s, nil
}
