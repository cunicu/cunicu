package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/device"
	"github.com/stv0g/cunicu/pkg/wg"

	errorsx "github.com/stv0g/cunicu/pkg/errors"
)

type wgProvider struct {
	path   string
	order  []string
	logger *zap.Logger
}

func WireGuardProvider() *wgProvider {
	path := os.Getenv("WG_CONFIG_PATH")
	if path == "" {
		path = wg.ConfigPath
	}

	return &wgProvider{
		path: path,

		logger: zap.L().Named("config.wg"),
	}
}

func (p *wgProvider) Read() (map[string]interface{}, error) {
	des, err := os.ReadDir(p.path)
	if err != nil {
		return nil, fmt.Errorf("failed to list config files in '%s': %w", p.path, err)
	}

	m := map[string]any{}

	p.order = []string{}

	for _, de := range des {
		cfg := filepath.Join(p.path, de.Name())
		filename := path.Base(cfg)
		extension := path.Ext(filename)
		name := strings.TrimSuffix(filename, extension)

		if extension != ".conf" {
			p.logger.Warn("Ignoring non-configuration file", zap.String("config_file", cfg))
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

func (p *wgProvider) ReadBytes() ([]byte, error) {
	return nil, errors.New("this provider does not support parsers")
}

func (p *wgProvider) Order() []string {
	slices.Sort(p.order)

	return p.order
}

type wgParser struct {
}

func (p *wgParser) Unmarshal(data []byte) (map[string]interface{}, error) {
	c, err := wg.ParseConfig(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %s", err)
	}

	s, err := NewInterfaceSettingsFromConfig(c)
	if err != nil {
		return nil, err
	}

	return Map(s, "koanf"), nil
}

func (p *wgParser) Marshal(map[string]interface{}) ([]byte, error) {
	return nil, errorsx.ErrNotSupported
}

func NewInterfaceSettingsFromConfig(c *wg.Config) (*InterfaceSettings, error) {
	s := &InterfaceSettings{
		WireGuard: WireGuardSettings{
			ListenPort: c.ListenPort,
			Peers:      map[string]WireGuardPeerSettings{},
		},
		AutoConfig: AutoConfigSettings{
			Addresses: c.Address,
			DNS:       c.DNS,
		},
	}

	if c.PrivateKey != nil {
		s.WireGuard.PrivateKey = crypto.Key(*c.PrivateKey)
	}

	if c.FirewallMark != nil {
		s.WireGuard.FirewallMark = *c.FirewallMark
	}

	if c.MTU != nil {
		s.AutoConfig.MTU = *c.MTU
	}

	if c.Table != nil {
		var err error

		s.RouteSync.Table, err = device.Table(*c.Table)
		if err != nil {
			return nil, fmt.Errorf("failed to parse routing table '%s': %w", *c.Table, err)
		}
	}

	// TODO: Add exec hooks for wg-quick PreUp, PostUp, PreDown, PostDown hooks

	for i, p := range c.Peers {
		wgps := WireGuardPeerSettings{
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

		s.WireGuard.Peers[name] = wgps
	}

	return s, nil
}
