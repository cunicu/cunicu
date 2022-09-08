package config

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"sync"

	"github.com/pion/ice/v2"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func (c *Config) Lookup(name string) error {
	g := errgroup.Group{}

	g.Go(func() error { return c.lookupTXT(name) })
	g.Go(func() error { return c.lookupSRV(name) })

	return g.Wait()
}

func (c *Config) lookupTXT(name string) error {
	rr, err := net.LookupTXT(name)
	if err != nil {
		return err
	}

	var re = regexp.MustCompile(`^(?m)cunicu-(.+?)=(.*)$`)

	c.logger.Debug("TXT records found", zap.Any("records", rr))

	rrs := map[string][]string{}
	for _, r := range rr {
		if m := re.FindStringSubmatch(r); m != nil {
			key := m[1]
			value := m[2]

			if _, ok := rrs[key]; !ok {
				rrs[key] = []string{value}
			} else {
				rrs[key] = append(rrs[key], value)
			}
		}
	}

	txtSettingMap := map[string]string{
		"community":    "peer_disc.community",
		"ice-username": "endpoint_disc.ice.username",
		"ice-password": "endpoint_disc.ice.password",
	}

	for txtName, settingName := range txtSettingMap {
		if values, ok := rrs[txtName]; ok {
			if len(values) > 1 {
				c.logger.Warn(fmt.Sprintf("Ignoring TXT record 'cunicu-%s' as there are more than one records with this prefix", txtName))
			} else {
				// We use SetDefault here as we do not want to overwrite user-provided settings with settings gathered via DNS
				c.SetDefault(settingName, values[0])
			}
		}
	}

	if backends, ok := rrs["backend"]; ok {
		c.Set("backends", backends)
	}

	if configFiles, ok := rrs["config"]; ok {
		for _, configFile := range configFiles {
			if u, err := url.Parse(configFile); err == nil {
				if err := c.MergeRemoteConfig(u); err != nil {
					return fmt.Errorf("failed to fetch config file from URL in cunicu-config TXT record: %s", err)
				}
			} else {
				return fmt.Errorf("failed to parse URL of config-file in cunicu-config TXT record: %s", err)
			}
		}
	}

	return nil
}

func (c *Config) lookupSRV(name string) error {
	svcs := map[string][]string{
		"stun":  {"udp"},
		"stuns": {"tcp"},
		"turn":  {"udp", "tcp"},
		"turns": {"tcp"},
	}

	urls := []string{}
	mu := sync.Mutex{}

	g := errgroup.Group{}

	reqs := 0
	for svc, protos := range svcs {
		for _, proto := range protos {
			reqs++
			s := svc
			p := proto
			g.Go(func() error {
				us, err := lookupICEUrlSRV(name, s, p)
				if err != nil {
					return err
				}

				mu.Lock()
				defer mu.Unlock()

				urls = append(urls, us...)

				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		return err
	}

	// We use SetDefault here as we do not want to overwrite user-provided settings with settings gathered via DNS
	c.SetDefault("endpoint_disc.ice.urls", urls)

	return nil
}

func lookupICEUrlSRV(name, svc, proto string) ([]string, error) {
	_, addrs, err := net.LookupSRV(svc, proto, name)
	if err != nil {
		return nil, err
	}

	urls := []string{}
	for _, addr := range addrs {
		url := ice.URL{
			Scheme: ice.NewSchemeType(svc),
			Host:   addr.Target,
			Port:   int(addr.Port),
			Proto:  ice.NewProtoType(proto),
		}
		urls = append(urls, url.String())
	}

	return urls, nil
}
