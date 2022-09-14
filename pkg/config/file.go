package config

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/stv0g/cunicu/pkg/util/buildinfo"
	"gopkg.in/yaml.v3"

	kyaml "github.com/knadh/koanf/parsers/yaml"
)

type fileProvider struct {
	InterfaceOrder []string

	url *url.URL
}

func YAMLFileProvider(u *url.URL) *fileProvider {
	return &fileProvider{
		url: u,
	}
}

func (p *fileProvider) ReadBytes() ([]byte, error) {
	return nil, errors.New("this provider requires no parser")
}

func (p *fileProvider) readBytes() ([]byte, error) {
	var err error
	var out []byte

	switch p.url.Scheme {
	case "http", "https":
		out, err = p.readBytesRemote()
	case "":
		out, err = p.readBytesLocal()
	default:
		err = fmt.Errorf("unsupported URL scheme: %s", p.url.Scheme)
	}
	if err != nil {
		return nil, err
	}

	p.InterfaceOrder, err = ExtractInterfaceOrder(out)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (p *fileProvider) Read() (map[string]interface{}, error) {
	buf, err := p.readBytes()
	if err != nil {
		return nil, err
	}

	return kyaml.Parser().Unmarshal(buf)
}

func (p *fileProvider) readBytesLocal() ([]byte, error) {
	f, err := os.Open(p.url.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file '%s': %w", p.url.Path, err)
	}

	defer f.Close()

	return io.ReadAll(f)
}

func (p *fileProvider) readBytesRemote() ([]byte, error) {
	if p.url.Scheme != "https" {
		host, _, err := net.SplitHostPort(p.url.Host)
		if err != nil {
			return nil, fmt.Errorf("failed to split host:port: %w", err)
		} else if host != "localhost" && host != "127.0.0.1" && host != "::1" && host != "[::1]" {
			return nil, errors.New("remote configuration must be provided via HTTPS")
		}
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req := &http.Request{
		Method: "GET",
		URL:    p.url,
		Header: http.Header{},
	}

	req.Header.Set("User-Agent", buildinfo.UserAgent())

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", p.url, err)
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch: %s: %s", p.url, resp.Status)
	}

	return io.ReadAll(resp.Body)
}

type stringMapSlice []string

func (keys *stringMapSlice) UnmarshalYAML(v *yaml.Node) error {
	if v.Kind != yaml.MappingNode {
		return fmt.Errorf("pipeline must contain YAML mapping, has %v", v.Kind)
	}

	*keys = make([]string, len(v.Content)/2)
	for i := 0; i < len(v.Content); i += 2 {
		if err := v.Content[i].Decode(&(*keys)[i/2]); err != nil {
			return err
		}
	}

	return nil
}

func ExtractInterfaceOrder(buf []byte) ([]string, error) {
	var s struct {
		Interfaces stringMapSlice `yaml:"interfaces,omitempty"`
	}

	if err := yaml.Unmarshal(buf, &s); err != nil {
		return nil, err
	}

	return s.Interfaces, nil
}
