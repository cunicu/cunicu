package config

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"sync"

	"github.com/knadh/koanf/maps"
	"github.com/pion/ice/v2"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type lookupProvider struct {
	Domain string

	Files    []string
	settings map[string]any

	mu     sync.Mutex
	logger *zap.Logger
}

func LookupProvider(domain string) *lookupProvider {
	logger := zap.L().Named("lookup")

	return &lookupProvider{
		Domain: domain,

		settings: map[string]any{},
		logger:   logger,
	}
}

func (p *lookupProvider) set(key string, value any) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.settings[key] = value
}

func (p *lookupProvider) ReadBytes() ([]byte, error) {
	return nil, errors.New("this provider requires no parser")
}

func (p *lookupProvider) Read() (map[string]any, error) {
	g := errgroup.Group{}

	g.Go(func() error { return p.lookupTXT() })
	g.Go(func() error { return p.lookupSRV() })

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return maps.Unflatten(p.settings, "."), nil
}

func (p *lookupProvider) lookupTXT() error {
	rr, err := net.LookupTXT(p.Domain)
	if err != nil {
		return err
	}

	var re = regexp.MustCompile(`^(?m)cunicu-(.+?)=(.*)$`)

	p.logger.Debug("TXT records found", zap.Any("records", rr))

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
		"community":    "pdisc.community",
		"ice-username": "epdisc.ice.username",
		"ice-password": "epdisc.ice.password",
	}

	for txtName, settingName := range txtSettingMap {
		if values, ok := rrs[txtName]; ok {
			if len(values) > 1 {
				p.logger.Warn(fmt.Sprintf("Ignoring TXT record 'cunicu-%s' as there are more than one records with this prefix", txtName))
			} else {
				p.set(settingName, values[0])
			}
		}
	}

	if backends, ok := rrs["backend"]; ok {
		p.set("backends", backends)
	}

	if files, ok := rrs["config"]; ok {
		p.Files = append(p.Files, files...)
	}

	return nil
}

func (p *lookupProvider) lookupSRV() error {
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
			q := proto
			g.Go(func() error {
				us, err := lookupICEUrlSRV(p.Domain, s, q)
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
	p.set("epdisc.ice.urls", urls)

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
