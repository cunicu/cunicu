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
	PrivateKey *string
	ListenPort *int
	FwMark     *int `ini:",omitempty"`

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
	PublicKey           string
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

func DumpConfig(wr io.Writer, cfg *Config) error {
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
	var cfg = &Config{
		Config: wgtypes.Config{
			Peers: []wgtypes.PeerConfig{},
		},
	}

	iniFile, err := ini.LoadSources(ini.LoadOptions{
		AllowNonUniqueSections: true,
		AllowShadows:           true,
	}, rd)
	if err != nil {
		return nil, err
	}

	iniCfg := &config{}
	if err := iniFile.StrictMapTo(iniCfg); err != nil {
		return nil, fmt.Errorf("failed to parse Interface section: %s", err)
	}

	cfg.ListenPort = iniCfg.Interface.ListenPort
	cfg.FirewallMark = iniCfg.Interface.FwMark

	if iniCfg.Interface.PrivateKey != nil {
		pk, err := wgtypes.ParseKey(*iniCfg.Interface.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse key: %w", err)
		}

		cfg.PrivateKey = &pk
	}

	// wg-quick settings
	cfg.MTU = iniCfg.Interface.MTU
	cfg.Table = iniCfg.Interface.Table
	cfg.PreUp = iniCfg.Interface.PreUp
	cfg.PreDown = iniCfg.Interface.PreDown
	cfg.PostUp = iniCfg.Interface.PostUp
	cfg.PostDown = iniCfg.Interface.PostDown
	cfg.SaveConfig = iniCfg.Interface.SaveConfig

	if iniCfg.Interface.Address != nil {
		if cfg.Address, err = parseCIDRs(iniCfg.Interface.Address, true); err != nil {
			return nil, err
		}
	}

	if iniCfg.Interface.DNS != nil {
		if cfg.DNS, err = parseIPs(iniCfg.Interface.DNS); err != nil {
			return nil, err
		}
	}

	for _, pSection := range iniCfg.Peers {

		peer := wgtypes.PeerConfig{
			AllowedIPs: []net.IPNet{},
		}

		peer.PublicKey, err = wgtypes.ParseKey(pSection.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse key: %w", err)
		}

		if pSection.PresharedKey != nil {
			psk, err := wgtypes.ParseKey(*pSection.PresharedKey)
			if err != nil {
				return nil, fmt.Errorf("failed to parse key: %w", err)
			}

			peer.PresharedKey = &psk
		}

		if pSection.PersistentKeepalive != nil {
			ka := time.Duration(*pSection.PersistentKeepalive) * time.Second
			peer.PersistentKeepaliveInterval = &ka
		}

		if pSection.Endpoint != nil {
			addr, err := net.ResolveUDPAddr("udp", *pSection.Endpoint)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve endpoint %s: %w", *pSection.Endpoint, err)
			}
			peer.Endpoint = addr
		}

		if pSection.AllowedIPs != nil {
			if peer.AllowedIPs, err = parseCIDRs(pSection.AllowedIPs, false); err != nil {
				return nil, fmt.Errorf("failed to parse allowed ips: %w", err)
			}
		}

		cfg.Peers = append(cfg.Peers, peer)
	}

	return cfg, nil
}
