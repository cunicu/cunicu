// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package nodes

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"time"

	g "cunicu.li/gont/v2/pkg"
	copt "cunicu.li/gont/v2/pkg/options/cmd"
	"github.com/pion/stun/v3"
	"go.uber.org/zap"

	"cunicu.li/cunicu/pkg/log"
)

var errTimeout = errors.New("timed out")

const (
	relayUsername = "user1"
	relayPassword = "password1"
)

type CoturnNode struct {
	*g.Host

	Command *g.Cmd

	Config map[string]string

	logger *log.Logger
}

func NewCoturnNode(n *g.Network, name string, opts ...g.Option) (*CoturnNode, error) {
	h, err := n.AddHost(name, opts...)
	if err != nil {
		return nil, err
	}

	logPath := fmt.Sprintf("%s.log", name)

	t := &CoturnNode{
		Host: h,
		Config: map[string]string{
			"verbose":                  "",
			"no-tls":                   "",
			"no-dtls":                  "",
			"lt-cred-mech":             "",
			"simple-log":               "",
			"no-stdout-log":            "",
			"log-file":                 logPath,
			"new-log-timestamp":        "",
			"new-log-timestamp-format": "%H:%M:%S",
			"listening-port":           strconv.Itoa(stun.DefaultPort),
			"realm":                    "cunicu",
			"cli-password":             "cunicu",
			"user":                     fmt.Sprintf("%s:%s", relayUsername, relayPassword),
		},
		logger: log.Global.Named("node.relay").With(zap.String("node", name)),
	}

	return t, nil
}

func (c *CoturnNode) Start(_, dir string, extraArgs ...any) error {
	var err error

	// Delete previous log file
	_ = os.Remove(c.Config["log-file"])

	args := []any{
		copt.Dir(dir),
		"-n",
	}
	args = append(args, extraArgs...)

	for key, value := range c.Config {
		opt := fmt.Sprintf("--%s", key)
		if value != "" {
			opt += fmt.Sprintf("=%s", value)
		}

		args = append(args, opt)
	}

	if c.Command, err = c.Host.Start("turnserver", args...); err != nil {
		return fmt.Errorf("failed to start turnserver: %w", err)
	}

	if err := c.WaitReady(); err != nil {
		return fmt.Errorf("failed to start turn server: %w", err)
	}

	return nil
}

func (c *CoturnNode) Stop() error {
	if c.Command == nil || c.Command.Process == nil {
		return nil
	}

	c.logger.Info("Stopping relay node")

	if err := GracefullyTerminate(c.Command); err != nil {
		// Coturn exits with exit code 143 (SIGTERM received)
		exitErr := &exec.ExitError{}
		if ok := errors.As(err, &exitErr); ok && exitErr.ExitCode() == 143 {
			return nil
		}
	}

	return nil
}

func (c *CoturnNode) Close() error {
	return c.Stop()
}

func (c *CoturnNode) isReachable() bool {
	hostPort := fmt.Sprintf("[%s]:%d", net.IPv6loopback, stun.DefaultPort)

	return c.RunFunc(func() error {
		conn, err := net.Dial("tcp6", hostPort)
		if err != nil {
			return err
		}

		return conn.Close()
	}) == nil
}

func (c *CoturnNode) WaitReady() error {
	for tries := 1000; !c.isReachable(); tries-- {
		if tries == 0 {
			return errTimeout
		}

		time.Sleep(20 * time.Millisecond)
	}

	return nil
}

func (c *CoturnNode) URLs() []url.URL {
	host := c.Name()
	hostPort := fmt.Sprintf("%s:%d", host, stun.DefaultPort)
	userHostPort := fmt.Sprintf("%s:%s@%s", relayUsername, relayPassword, hostPort)

	return []url.URL{
		{
			Scheme: "stun",
			Opaque: hostPort,
		},
		{
			Scheme:   "turn",
			Opaque:   userHostPort,
			RawQuery: "transport=udp",
		},
		{
			Scheme:   "turn",
			Opaque:   userHostPort,
			RawQuery: "transport=tcp",
		},
	}
}

// Options
func (c *CoturnNode) Apply(i *g.Interface) {
	i.Node = c.Host
}
