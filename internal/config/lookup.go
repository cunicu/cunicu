package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"sync"

	"github.com/pion/ice/v2"
	"go.uber.org/zap"
)

func (c *Config) Lookup(name string) error {
	wg := &sync.WaitGroup{}
	wg.Add(2)

	haveError := false
	errs := make(chan error)

	go c.lookupTXT(name, errs, wg)
	go c.lookupSRV(name, errs, wg)
	go func() {
		for err := range errs {
			c.logger.Error("Failed to load autoconfig", zap.Error(err))
			haveError = true
		}
	}()
	wg.Wait()

	if haveError {
		return errors.New("failed to lookup configuration")
	}

	return nil
}

func (c *Config) lookupTXT(name string, errs chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	rr, err := net.LookupTXT(name)
	if err != nil {
		errs <- err
		return
	}

	var re = regexp.MustCompile(`^(?m)wice-(.+?)=(.*)$`)

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
		"community":    "community",
		"ice-username": "ice.username",
		"ice-password": "ice.password",
	}

	for txtName, settingName := range txtSettingMap {
		c.setSingleTxtRecord(rrs, txtName, settingName)
	}

	if backends, ok := rrs["backend"]; ok {
		c.Set("backends", backends)
	}

	if configFiles, ok := rrs["config"]; ok {
		for _, configFile := range configFiles {
			if u, err := url.Parse(configFile); err == nil {
				if err := c.MergeRemoteConfig(u); err != nil {
					c.logger.Warn("Ignoring invalid URL in wice-config TXT record", zap.Error(err))
				}
			} else {
				c.logger.Warn("Ignoring invalid URL in wice-config TXT record", zap.Error(err))
				return
			}
		}
	}
}

func (c *Config) lookupSRV(name string, errs chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	svcs := map[string][]string{
		"stun":  {"udp"},
		"stuns": {"tcp"},
		"turn":  {"udp", "tcp"},
		"turns": {"tcp"},
	}

	urls := []string{}
	mu := &sync.Mutex{}
	wg2 := &sync.WaitGroup{}
	for svc, protos := range svcs {
		for _, proto := range protos {
			wg2.Add(1)
			go func(svc, proto string) {
				defer wg2.Done()
				if us, err := lookupICEUrlSRV(name, svc, proto); err == nil {
					mu.Lock()
					urls = append(urls, us...)
					mu.Unlock()
				}
			}(svc, proto)
		}
	}

	wg2.Wait()

	// We use SetDefault here as we dont want to overwrite user-provided settings with settings gathered via DNS
	c.SetDefault("ice.urls", urls)
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

func (c *Config) setSingleTxtRecord(txtRecords map[string][]string, txtName string, settingName string) {
	if values, ok := txtRecords[txtName]; ok {
		if len(values) > 1 {
			c.logger.Warn(fmt.Sprintf("Ignoring TXT record 'wice-%s' as there are more than once records with this prefix", txtName))
		} else {
			// We use SetDefault here as we dont want to overwrite user-provided settings with settings gathered via DNS
			c.SetDefault(settingName, values[0])
		}
	}
}
