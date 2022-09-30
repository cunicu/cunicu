package config

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"sync"
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/maps"
	"github.com/miekg/dns"
	"github.com/pion/ice/v2"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type lookupProvider struct {
	domain     string
	lastSerial int
	files      []string
	settings   map[string]any

	mu     sync.Mutex
	logger *zap.Logger
}

func LookupProvider(domain string) *lookupProvider {
	logger := zap.L().Named("lookup")

	return &lookupProvider{
		domain: domain,

		settings: map[string]any{},
		logger:   logger,
	}
}

func (p *lookupProvider) ReadBytes() ([]byte, error) {
	return nil, errors.New("this provider requires no parser")
}

func (p *lookupProvider) Read() (map[string]any, error) {
	g := errgroup.Group{}

	g.Go(func() error { return p.lookupTXT(context.Background()) })
	g.Go(func() error { return p.lookupSRV(context.Background()) })

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return maps.Unflatten(p.settings, "."), nil
}

func (p *lookupProvider) Watch(cb func(event interface{}, err error)) error {
	go func() {
		t := time.NewTicker(5 * time.Minute)
		for range t.C {
			serial, err := p.lookupSerial(context.Background())
			if err != nil {
				p.logger.Error("Failed to lookup zones SOA serial", zap.Error(err))
				continue
			}

			if serial != p.lastSerial {
				p.lastSerial = serial
				cb(nil, nil)
			}
		}
	}()

	return nil
}

func (p *lookupProvider) Version() any {
	var err error

	if p.lastSerial, err = p.lookupSerial(context.Background()); err != nil {
		return nil
	}

	return p.lastSerial
}

func (p *lookupProvider) SubProviders() []koanf.Provider {
	ps := []koanf.Provider{}

	for _, f := range p.files {
		u, err := url.Parse(f)
		if err != nil {
			p.logger.Warn("failed to parse URL for configuration file", zap.Error(err))
		} else {
			ps = append(ps, RemoteFileProvider(u))
		}
	}

	return ps
}

func (p *lookupProvider) set(key string, value any) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.settings[key] = value
}

func (p *lookupProvider) lookupSerial(ctx context.Context) (int, error) {
	var err error
	var conn *dns.Conn

	cfg, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		return -1, fmt.Errorf("failed to load resolver configuration: %w", err)
	}

	addr := net.JoinHostPort(cfg.Servers[0], cfg.Port)

	if res := net.DefaultResolver; res.PreferGo {
		dial := res.Dial
		if dial == nil {
			var d net.Dialer
			dial = d.DialContext
		}

		connNet, err := dial(ctx, "udp", addr)
		if err != nil {
			return -1, fmt.Errorf("failed to connect to %s: %w", addr, err)
		}

		conn = &dns.Conn{
			Conn: connNet,
		}
	} else {
		client := dns.Client{}
		conn, err = client.DialContext(ctx, addr)
		if err != nil {
			return -1, fmt.Errorf("failed to connect to %s: %w", addr, err)
		}
	}

	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(p.domain), dns.TypeSOA)

	conn.WriteMsg(msg)
	resp, err := conn.ReadMsg()
	if err != nil {
		return -1, fmt.Errorf("failed to read response: %w", err)
	}

	if err := conn.Close(); err != nil {
		return -1, fmt.Errorf("failed to close connection: %w", err)
	}

	for _, ans := range resp.Answer {
		if s, ok := ans.(*dns.SOA); ok {
			return int(s.Serial), nil
		}
	}

	return -1, errors.New("failed to find SOA record")
}

func (p *lookupProvider) lookupTXT(ctx context.Context) error {
	rr, err := net.LookupTXT(p.domain)
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

	if fs, ok := rrs["config"]; ok {
		p.files = append(p.files, fs...)
	}

	return nil
}

func (p *lookupProvider) lookupSRV(ctx context.Context) error {
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
				us, err := lookupICEUrlSRV(p.domain, s, q)
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
