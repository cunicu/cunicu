package config

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/knadh/koanf/providers/file"
	"github.com/stv0g/cunicu/pkg/util/buildinfo"
)

type remoteFileProvider struct {
	url          *url.URL
	etag         string
	lastModified time.Time
	order        []string
}

func RemoteFileProvider(u *url.URL) *remoteFileProvider {
	return &remoteFileProvider{
		url: u,
	}
}

func (p *remoteFileProvider) Read() (map[string]interface{}, error) {
	return nil, errors.New("this provider does not support parsers")
}

func (p *remoteFileProvider) ReadBytes() ([]byte, error) {
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

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	p.order, err = ExtractInterfaceOrder(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to get interface order: %w", err)
	}

	p.etag = resp.Header.Get("Etag")

	if lm := resp.Header.Get("Last-modified"); lm != "" {
		p.lastModified, err = time.Parse(http.TimeFormat, lm)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Last-Modified header: %w", err)
		}
	}

	return buf, nil
}

func (p *remoteFileProvider) Order() []string {
	return p.order
}

func (p *remoteFileProvider) Version() any {
	if p.etag != "" {
		return p.etag
	}

	if !p.lastModified.IsZero() {
		return p.lastModified.Unix()
	}

	return nil
}

func (p *remoteFileProvider) hasChanged() (bool, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req := &http.Request{
		Method: "HEAD",
		URL:    p.url,
		Header: http.Header{},
	}

	req.Header.Set("User-Agent", buildinfo.UserAgent())

	if p.etag != "" {
		req.Header.Set("If-None-Match", fmt.Sprintf("\"%s\"", p.etag))
	}

	if !p.lastModified.IsZero() {
		req.Header.Set("If-Modified-Since", p.lastModified.Format(http.TimeFormat))
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to fetch %s: %w", p.url, err)
	} else if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("failed to fetch: %s: %s", p.url, resp.Status)
	}

	return resp.StatusCode == 200, nil
}

type localFileProvider struct {
	*file.File

	order []string
}

func LocalFileProvider(u *url.URL) *localFileProvider {
	return &localFileProvider{
		File: file.Provider(u.Path),
	}
}

func (p *localFileProvider) ReadBytes() ([]byte, error) {
	buf, err := p.File.ReadBytes()

	if err == nil {
		p.order, err = ExtractInterfaceOrder(buf)
	}

	return buf, err
}

func (p *localFileProvider) Order() []string {
	return p.order
}
