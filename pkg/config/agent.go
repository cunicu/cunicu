package config

import (
	"github.com/pion/ice/v2"
)

func (c *Config) AgentConfig() (*ice.AgentConfig, error) {
	cfg := &ice.AgentConfig{
		InsecureSkipVerify: c.EndpointDisc.ICE.InsecureSkipVerify,
		Lite:               c.EndpointDisc.ICE.Lite,
		PortMin:            uint16(c.EndpointDisc.ICE.Port.Min),
		PortMax:            uint16(c.EndpointDisc.ICE.Port.Max),
	}

	cfg.InterfaceFilter = func(name string) bool {
		return c.EndpointDisc.ICE.InterfaceFilter.MatchString(name)
	}

	// ICE URLs
	cfg.Urls = []*ice.URL{}
	for _, u := range c.EndpointDisc.ICE.URLs {
		p := u.URL

		p.Username = c.EndpointDisc.ICE.Username
		p.Password = c.EndpointDisc.ICE.Password

		cfg.Urls = append(cfg.Urls, &p)
	}

	if len(c.EndpointDisc.ICE.NAT1to1IPs) > 0 {
		cfg.NAT1To1IPs = c.EndpointDisc.ICE.NAT1to1IPs
	}

	if mbr := uint16(c.EndpointDisc.ICE.MaxBindingRequests); mbr > 0 {
		cfg.MaxBindingRequests = &mbr
	}

	if c.EndpointDisc.ICE.MDNS {
		cfg.MulticastDNSMode = ice.MulticastDNSModeQueryAndGather
	}

	if to := c.EndpointDisc.ICE.DisconnectedTimeout; to > 0 {
		cfg.DisconnectedTimeout = &to
	}

	if to := c.EndpointDisc.ICE.FailedTimeout; to > 0 {
		cfg.FailedTimeout = &to
	}

	if to := c.EndpointDisc.ICE.KeepaliveInterval; to > 0 {
		cfg.KeepaliveInterval = &to
	}

	if to := c.EndpointDisc.ICE.CheckInterval; to > 0 {
		cfg.CheckInterval = &to
	}

	if len(c.EndpointDisc.ICE.CandidateTypes) > 0 {
		cfg.CandidateTypes = []ice.CandidateType{}
		for _, t := range c.EndpointDisc.ICE.CandidateTypes {
			cfg.CandidateTypes = append(cfg.CandidateTypes, t.CandidateType)
		}
	}

	if len(c.EndpointDisc.ICE.NetworkTypes) > 0 {
		cfg.NetworkTypes = []ice.NetworkType{}
		for _, t := range c.EndpointDisc.ICE.NetworkTypes {
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
