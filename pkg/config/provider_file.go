// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"time"

	"github.com/knadh/koanf/providers/file"

	"github.com/stv0g/cunicu/pkg/buildinfo"
)

var (
	errInsecureRemoteConfig = errors.New("remote configuration must be provided via HTTPS")
	errFailedToFetch        = errors.New("failed to fetch")
	errInsecurePermissions  = errors.New("insecure permissions on configuration file")
)

type RemoteFileProvider struct {
	url          *url.URL
	etag         string
	lastModified time.Time
	order        []string
}

func NewRemoteFileProvider(u *url.URL) *RemoteFileProvider {
	return &RemoteFileProvider{
		url: u,
	}
}

func (p *RemoteFileProvider) Read() (map[string]interface{}, error) {
	return nil, errNotImplemented
}

func (p *RemoteFileProvider) ReadBytes() ([]byte, error) {
	if p.url.Scheme != "https" {
		host, _, err := net.SplitHostPort(p.url.Host)
		if err != nil {
			return nil, fmt.Errorf("failed to split host:port: %w", err)
		} else if host != "localhost" && host != "127.0.0.1" && host != "::1" && host != "[::1]" {
			return nil, errInsecureRemoteConfig
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
		return nil, fmt.Errorf("%w: %s: %s", errFailedToFetch, p.url, resp.Status)
	}
	defer resp.Body.Close()

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

func (p *RemoteFileProvider) Order() []string {
	return p.order
}

func (p *RemoteFileProvider) Version() any {
	if _, err := p.hasChanged(); err != nil {
		return nil
	}

	if p.etag != "" {
		return p.etag
	}

	if !p.lastModified.IsZero() {
		return p.lastModified.Unix()
	}

	return nil
}

func (p *RemoteFileProvider) hasChanged() (bool, error) {
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
		return false, fmt.Errorf("%w %s: %w", errFailedToFetch, p.url, err)
	} else if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("%w: %s: %s", errFailedToFetch, p.url, resp.Status)
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200, nil
}

type LocalFileProvider struct {
	*file.File
	path string

	order []string

	allowInsecure bool
}

func NewLocalFileProvider(path string) *LocalFileProvider {
	return &LocalFileProvider{
		File:          file.Provider(path),
		path:          path,
		allowInsecure: os.Getenv("CUNICU_CONFIG_ALLOW_INSECURE") == "true" || runtime.GOOS == "windows",
	}
}

func (p *LocalFileProvider) ReadBytes() ([]byte, error) {
	if !p.allowInsecure {
		fi, err := os.Stat(p.path)
		if err != nil {
			return nil, err
		}

		if perm := fi.Mode().Perm(); perm&0o7 != 0 {
			return nil, fmt.Errorf("%w: %s", errInsecurePermissions, p.path)
		}
	}

	buf, err := p.File.ReadBytes()

	if err == nil {
		p.order, err = ExtractInterfaceOrder(buf)
	}

	return buf, err
}

func (p *LocalFileProvider) Order() []string {
	return p.order
}

var windowsDriveLetterRegexp = regexp.MustCompile(`(?i)^[a-z]$`)

func isWindowsDriveLetter(s string) bool {
	return windowsDriveLetterRegexp.MatchString(s)
}
