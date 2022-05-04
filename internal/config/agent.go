package config

import (
	"fmt"
	"io"
	"regexp"

	"github.com/pion/ice/v2"
	"gopkg.in/yaml.v3"
	icex "riasc.eu/wice/internal/ice"
)

func (c *Config) AgentConfig() (*ice.AgentConfig, error) {
	cfg := &ice.AgentConfig{
		InsecureSkipVerify: c.GetBool("ice.insecure_skip_verify"),
		Lite:               c.GetBool("ice.lite"),
		PortMin:            uint16(c.GetUint("ice.port.min")),
		PortMax:            uint16(c.GetUint("ice.port.max")),
	}

	interfaceFilterRegex, err := regexp.Compile(c.GetString("ice.interface_filter"))
	if err != nil {
		return nil, fmt.Errorf("invalid ice.interface_filter config: %w", err)
	}

	cfg.InterfaceFilter = func(name string) bool {
		return interfaceFilterRegex.Match([]byte(name))
	}

	// ICE URLS
	cfg.Urls = []*ice.URL{}
	for _, u := range c.GetStringSlice("ice.urls") {
		up, err := ice.ParseURL(u)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ice.url: %s: %w", u, err)
		}

		cfg.Urls = append(cfg.Urls, up)
	}

	// Add default STUN/TURN servers
	// Set ICE credentials
	u := c.GetString("ice.username")
	p := c.GetString("ice.password")
	for _, q := range cfg.Urls {
		if u != "" {
			q.Username = u
		}

		if p != "" {
			q.Password = p
		}
	}

	if c.IsSet("ice.nat_1to1_ips") {
		cfg.NAT1To1IPs = c.GetStringSlice("ice.nat_1to1_ips")
	}

	if c.IsSet("ice.max_binding_requests") {
		i := uint16(c.GetInt("ice.max_binding_requests"))
		cfg.MaxBindingRequests = &i
	}

	if c.GetBool("ice.mdns") {
		cfg.MulticastDNSMode = ice.MulticastDNSModeQueryAndGather
	}

	if c.IsSet("ice.disconnected_timeout") {
		to := c.GetDuration("ice.disconnected_timeout")
		cfg.DisconnectedTimeout = &to
	}

	if c.IsSet("ice.failed_timeout") {
		to := c.GetDuration("ice.failed_timeout")
		cfg.FailedTimeout = &to
	}

	if c.IsSet("ice.keepalive_interval") {
		to := c.GetDuration("ice.keepalive_interval")
		cfg.KeepaliveInterval = &to
	}

	if c.IsSet("ice.check_interval") {
		to := c.GetDuration("ice.check_interval")
		cfg.CheckInterval = &to
	}

	// Filter candidate types
	candidateTypes := []ice.CandidateType{}
	for _, value := range c.GetStringSlice("ice.candidate_types") {
		ct, err := icex.CandidateTypeFromString(value)
		if err != nil {
			return nil, err
		}

		candidateTypes = append(candidateTypes, ct)
	}

	if len(candidateTypes) > 0 {
		cfg.CandidateTypes = candidateTypes
	}

	// Filter network types
	networkTypes := []ice.NetworkType{}
	for _, value := range c.GetStringSlice("ice.network_types") {
		ct, err := icex.NetworkTypeFromString(value)
		if err != nil {
			return nil, err
		}

		networkTypes = append(networkTypes, ct)
	}

	if len(networkTypes) > 0 {
		cfg.NetworkTypes = networkTypes
	}

	return cfg, nil
}

func (c *Config) Dump(wr io.Writer) error {
	enc := yaml.NewEncoder(wr)
	enc.SetIndent(2)
	return enc.Encode(c.AllSettings())
}
