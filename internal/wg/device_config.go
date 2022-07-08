package wg

import (
	"fmt"
	"io"
	"net"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"gopkg.in/ini.v1"
)

type Config struct {
	wgtypes.Config

	Address    []net.IPNet
	DNS        []net.IPAddr
	MTU        *int
	Table      *string
	PreUp      []string
	PreDown    []string
	PostUp     []string
	PostDown   []string
	SaveConfig *bool
}

type config struct {
	Interface interfaceConfig

	Peers []peerConfig `ini:"Peer,nonunique"`
}

type interfaceConfig struct {
	PrivateKey *string `ini:",omitempty"`
	ListenPort *int    `ini:",omitempty"`
	FwMark     *int    `ini:",omitempty"`

	// Settings for wg-quick
	Address    []string `ini:",omitempty" delim:", "`
	DNS        []string `ini:",omitempty" delim:", "`
	MTU        *int     `ini:",omitempty"`
	Table      *string  `ini:",omitempty"`
	PreUp      []string `ini:",omitempty,allowshadow"`
	PreDown    []string `ini:",omitempty,allowshadow"`
	PostUp     []string `ini:",omitempty,allowshadow"`
	PostDown   []string `ini:",omitempty,allowshadow"`
	SaveConfig *bool    `ini:",omitempty"`
}

type peerConfig struct {
	PublicKey           string   `ini:",omitempty"`
	PresharedKey        *string  `ini:",omitempty"`
	AllowedIPs          []string `ini:",omitempty" delim:", "`
	Endpoint            *string  `ini:",omitempty"`
	PersistentKeepalive *int     `ini:",omitempty"`
}

func parseCIDRs(nets []string, ip bool) ([]net.IPNet, error) {
	pn := []net.IPNet{}
	for _, ne := range nets {
		i, n, err := net.ParseCIDR(ne)
		if err != nil {
			return nil, fmt.Errorf("failed to parse network %s: %w", ne, err)
		}

		if ip {
			n.IP = i
		}

		pn = append(pn, *n)
	}

	return pn, nil
}

func parseIPs(ips []string) ([]net.IPAddr, error) {
	pips := []net.IPAddr{}
	for _, ip := range ips {
		i, err := net.ResolveIPAddr("ip", ip)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ip %s: %w", ip, err)
		}

		pips = append(pips, *i)
	}

	return pips, nil
}

func (cfg *Config) Dump(wr io.Writer) error {
	iniCfg := &config{
		Interface: interfaceConfig{
			ListenPort: cfg.ListenPort,
			FwMark:     cfg.FirewallMark,
		},
	}

	if cfg.PrivateKey != nil {
		pk := cfg.PrivateKey.String()
		iniCfg.Interface.PrivateKey = &pk
	}

	// wg-quick settings
	iniCfg.Interface.MTU = cfg.MTU
	iniCfg.Interface.SaveConfig = cfg.SaveConfig
	iniCfg.Interface.Table = cfg.Table
	iniCfg.Interface.PreUp = cfg.PreUp
	iniCfg.Interface.PreDown = cfg.PreDown
	iniCfg.Interface.PostUp = cfg.PostUp
	iniCfg.Interface.PostDown = cfg.PostDown
	iniCfg.Interface.SaveConfig = cfg.SaveConfig

	if len(cfg.Address) > 0 {
		iniCfg.Interface.Address = []string{}
		for _, addr := range cfg.Address {
			iniCfg.Interface.Address = append(iniCfg.Interface.Address, addr.String())
		}
	}

	if len(cfg.DNS) > 0 {
		iniCfg.Interface.DNS = []string{}
		for _, addr := range cfg.DNS {
			iniCfg.Interface.DNS = append(iniCfg.Interface.DNS, addr.String())
		}
	}

	for _, peer := range cfg.Peers {
		iniPeer := peerConfig{
			PublicKey: peer.PublicKey.String(),
		}

		if peer.PresharedKey != nil {
			psk := peer.PresharedKey.String()
			iniPeer.PresharedKey = &psk
		}

		if peer.Endpoint != nil {
			ep := peer.Endpoint.String()
			iniPeer.Endpoint = &ep
		}

		if peer.PersistentKeepaliveInterval != nil {
			ka := int(*peer.PersistentKeepaliveInterval / time.Second)
			iniPeer.PersistentKeepalive = &ka
		}

		if len(peer.AllowedIPs) > 0 {
			iniPeer.AllowedIPs = []string{}
			for _, aip := range peer.AllowedIPs {
				iniPeer.AllowedIPs = append(iniPeer.AllowedIPs, aip.String())
			}
		}

		iniCfg.Peers = append(iniCfg.Peers, iniPeer)
	}

	iniFile := ini.Empty(ini.LoadOptions{
		AllowShadows: true,
	})

	if err := iniFile.ReflectFrom(iniCfg); err != nil {
		return err
	}

	_, err := iniFile.WriteTo(wr)

	return err
}

func ParseConfig(rd io.Reader, name string) (*Config, error) {
	var err error

	iniFile, err := ini.LoadSources(ini.LoadOptions{
		AllowNonUniqueSections: true,
		AllowShadows:           true,
	}, rd)
	if err != nil {
		return nil, err
	}

	// We add a pseudo peer section just allow mapping via StrictMapTo if there are no peers configured
	fakePeer := !iniFile.HasSection("Peer")
	if fakePeer {
		iniFile.NewSection("Peer")
	}

	iniCfg := &config{}
	if err := iniFile.StrictMapTo(iniCfg); err != nil {
		return nil, fmt.Errorf("failed to parse Interface section: %s", err)
	}

	// Remove fake peer section again
	if fakePeer {
		iniCfg.Peers = nil
	}

	return iniCfg.Config()
}

func (c *config) Config() (*Config, error) {
	var err error
	var cfg = &Config{
		Config: wgtypes.Config{
			Peers:        []wgtypes.PeerConfig{},
			ListenPort:   c.Interface.ListenPort,
			FirewallMark: c.Interface.FwMark,
		},
	}

	if c.Interface.Address != nil {
		if cfg.Address, err = parseCIDRs(c.Interface.Address, true); err != nil {
			return nil, err
		}
	}

	if c.Interface.DNS != nil {
		if cfg.DNS, err = parseIPs(c.Interface.DNS); err != nil {
			return nil, err
		}
	}

	if c.Interface.PrivateKey != nil {
		pk, err := wgtypes.ParseKey(*c.Interface.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse key: %w", err)
		}

		cfg.PrivateKey = &pk
	}

	// wg-quick settings
	cfg.MTU = c.Interface.MTU
	cfg.Table = c.Interface.Table
	cfg.PreUp = c.Interface.PreUp
	cfg.PreDown = c.Interface.PreDown
	cfg.PostUp = c.Interface.PostUp
	cfg.PostDown = c.Interface.PostDown
	cfg.SaveConfig = c.Interface.SaveConfig

	for _, pSection := range c.Peers {
		peerCfg, err := pSection.Config()
		if err != nil {
			return nil, fmt.Errorf("failed to parse peer config: %w", err)
		}

		cfg.Peers = append(cfg.Peers, *peerCfg)
	}

	return cfg, nil
}

func (p *peerConfig) Config() (*wgtypes.PeerConfig, error) {
	var err error

	cfg := &wgtypes.PeerConfig{
		AllowedIPs: []net.IPNet{},
	}

	cfg.PublicKey, err = wgtypes.ParseKey(p.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse key: %w", err)
	}

	if p.PresharedKey != nil {
		psk, err := wgtypes.ParseKey(*p.PresharedKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse key: %w", err)
		}

		cfg.PresharedKey = &psk
	}

	if p.PersistentKeepalive != nil {
		ka := time.Duration(*p.PersistentKeepalive) * time.Second
		cfg.PersistentKeepaliveInterval = &ka
	}

	if p.Endpoint != nil {
		addr, err := net.ResolveUDPAddr("udp", *p.Endpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve endpoint %s: %w", *p.Endpoint, err)
		}
		cfg.Endpoint = addr
	}

	if p.AllowedIPs != nil {
		if cfg.AllowedIPs, err = parseCIDRs(p.AllowedIPs, false); err != nil {
			return nil, fmt.Errorf("failed to parse allowed ips: %w", err)
		}
	}

	return cfg, nil
}
