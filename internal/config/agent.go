package config

import (
	"github.com/pion/ice/v2"
)

func (c *Config) AgentConfig() (*ice.AgentConfig, error) {
	cfg := &ice.AgentConfig{
		InsecureSkipVerify: c.ICE.InsecureSkipVerify,
		Lite:               c.ICE.Lite,
		PortMin:            uint16(c.ICE.Port.Min),
		PortMax:            uint16(c.ICE.Port.Max),
	}

	cfg.InterfaceFilter = func(name string) bool {
		return c.ICE.InterfaceFilter.MatchString(name)
	}

	// ICE URLs
	cfg.Urls = []*ice.URL{}
	for _, u := range c.ICE.URLs {
		p := &u.URL

		// Set ICE credentials for TURN/TURNS servers
		if p.Scheme == ice.SchemeTypeTURN || p.Scheme == ice.SchemeTypeTURNS {
			p.Username = c.ICE.Username
			p.Password = c.ICE.Password
		}

		cfg.Urls = append(cfg.Urls, p)

	}

	if len(c.ICE.NAT1to1IPs) > 0 {
		cfg.NAT1To1IPs = c.ICE.NAT1to1IPs
	}

	if mbr := uint16(c.ICE.MaxBindingRequests); mbr > 0 {
		cfg.MaxBindingRequests = &mbr
	}

	if c.ICE.MDNS {
		cfg.MulticastDNSMode = ice.MulticastDNSModeQueryAndGather
	}

	if to := c.ICE.DisconnectedTimeout; to > 0 {
		cfg.DisconnectedTimeout = &to
	}

	if to := c.ICE.FailedTimeout; to > 0 {
		cfg.FailedTimeout = &to
	}

	if to := c.ICE.KeepaliveInterval; to > 0 {
		cfg.KeepaliveInterval = &to
	}

	if to := c.ICE.CheckInterval; to > 0 {
		cfg.CheckInterval = &to
	}

	if len(c.ICE.CandidateTypes) > 0 {
		cfg.CandidateTypes = []ice.CandidateType{}
		for _, t := range c.ICE.CandidateTypes {
			cfg.CandidateTypes = append(cfg.CandidateTypes, t.CandidateType)
		}
	}

	if len(c.ICE.NetworkTypes) > 0 {
		cfg.NetworkTypes = []ice.NetworkType{}
		for _, t := range c.ICE.NetworkTypes {
			cfg.NetworkTypes = append(cfg.NetworkTypes, t.NetworkType)
		}
	} else {
		cfg.NetworkTypes = []ice.NetworkType{
			ice.NetworkTypeTCP4,
			ice.NetworkTypeUDP4,
			ice.NetworkTypeTCP6,
			ice.NetworkTypeUDP6,
		}
	}

	return cfg, nil
}
