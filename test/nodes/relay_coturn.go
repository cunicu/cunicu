//go:build linux

package nodes

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/pion/ice/v2"
	g "github.com/stv0g/gont/pkg"
	"go.uber.org/zap"
)

const (
	stunPort = 3478
)

type CoturnNode struct {
	*g.Host

	Command *exec.Cmd

	Config map[string]string

	logger *zap.Logger
}

func NewCoturnNode(n *g.Network, name string, opts ...g.Option) (RelayNode, error) {
	h, err := n.AddHost(name, opts...)
	if err != nil {
		return nil, err
	}

	logPath := fmt.Sprintf("%s.log", name)

	t := &CoturnNode{
		Host: h,
		Config: map[string]string{
			"verbose":        "",
			"no-tls":         "",
			"no-dtls":        "",
			"lt-cred-mech":   "",
			"simple-log":     "",
			"no-stdout-log":  "",
			"log-file":       logPath,
			"listening-port": strconv.Itoa(stunPort),
			"realm":          "wice",
			"cli-password":   "wice",
		},
		logger: zap.L().Named("node.relay").With(zap.String("node", name)),
	}

	t.Config["user"] = fmt.Sprintf("%s:%s", t.Username(), t.Password())

	return t, nil
}

func (c *CoturnNode) Username() string {
	return "user1"
}

func (c *CoturnNode) Password() string {
	return "password1"
}

func (c *CoturnNode) Start(_, dir string, extraArgs ...any) error {
	var err error

	// Delete previous log file
	os.Remove(c.Config["log-file"])

	args := []any{
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

	if _, _, c.Command, err = c.StartWith("turnserver", nil, dir, args...); err != nil {
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

	return GracefullyTerminate(c.Command)
}

func (c *CoturnNode) Close() error {
	return c.Stop()
}

func (c *CoturnNode) isReachable() bool {
	hostPort := fmt.Sprintf("[%s]:%d", net.IPv6loopback, stunPort)

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
			return fmt.Errorf("timed out")
		}

		time.Sleep(20 * time.Millisecond)
	}

	return nil
}

func (c *CoturnNode) URLs() []*ice.URL {
	host := c.Name()

	return []*ice.URL{
		{
			Scheme: ice.SchemeTypeSTUN,
			Host:   host,
			Port:   stunPort,
			Proto:  ice.ProtoTypeUDP,
		},
		{
			Scheme: ice.SchemeTypeTURN,
			Host:   host,
			Port:   stunPort,
			Proto:  ice.ProtoTypeUDP,
		},
		{
			Scheme: ice.SchemeTypeTURN,
			Host:   host,
			Port:   stunPort,
			Proto:  ice.ProtoTypeTCP,
		},
	}
}

// Options
func (c *CoturnNode) Apply(i *g.Interface) {
	i.Node = c.Host
}
