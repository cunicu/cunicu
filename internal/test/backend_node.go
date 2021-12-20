//go:build linux
// +build linux

package test

import (
	"fmt"
	"net/url"
	"os/exec"

	g "github.com/stv0g/gont/pkg"
)

type Backend struct {
	*g.Host

	Port int

	Command *exec.Cmd
}

func NewBackend(m *g.Network, name string, opts ...g.Option) (*Backend, error) {
	h, err := m.AddHost(name, opts...)
	if err != nil {
		return nil, err
	}

	b := &Backend{
		Host: h,
		Port: 8080,
	}

	return b, nil
}

func (b *Backend) Start() error {
	stdout, stderr, cmd, err := b.StartGo("../cmd/wice-signal-http", "-port", b.Port)
	if err != nil {
		return err
	}

	b.Command = cmd

	FileWriter("logs/backend.log", stdout, stderr)

	return nil
}

func (b *Backend) Stop() error {
	if b.Command == nil || b.Command.Process == nil {
		return nil
	}

	return b.Command.Process.Kill()
}

func (b *Backend) Close() error {
	return b.Stop()
}

func (b *Backend) URL() *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", b.Name(), b.Port),
	}
}
