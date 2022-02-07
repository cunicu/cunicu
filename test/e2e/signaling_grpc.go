package e2e

import (
	"fmt"
	"net/url"
	"os/exec"

	g "github.com/stv0g/gont/pkg"
)

type GrpcSignalingNode struct {
	*g.Host

	port int

	Command *exec.Cmd
}

func NewGrpcSignalingNode(n *g.Network, name string) (SignalingNode, error) {
	h, err := n.AddHost(name)
	if err != nil {
		return nil, err
	}

	t := &GrpcSignalingNode{
		Host: h,
		port: 8080,
	}

	return t, nil
}

func (s *GrpcSignalingNode) Start(_ ...interface{}) error {
	var err error
	var logPath = fmt.Sprintf("logs/%s.log", s.Name())

	var args = []interface{}{
		"signal",
		"--log-level", "debug",
		"--log-file", logPath,
		"--listen", fmt.Sprintf(":%d", s.port),
	}

	cmd, err := buildWICE(s.Network())
	if err != nil {
		return fmt.Errorf("failed to build wice: %w", err)
	}

	if _, _, s.Command, err = s.Host.Start(cmd, args...); err != nil {
		return fmt.Errorf("failed to start wice: %w", err)
	}

	return nil
}

func (s *GrpcSignalingNode) Stop() error {
	if s.Command == nil || s.Command.Process == nil {
		return nil
	}

	return s.Command.Process.Kill()
}

func (s *GrpcSignalingNode) Close() error {
	return s.Stop()
}

func (s *GrpcSignalingNode) URL() url.URL {
	return url.URL{
		Scheme:   "grpc",
		Host:     fmt.Sprintf("127.0.0.1:%d", s.port),
		RawQuery: "insecure=true",
	}
}

// Options
func (s *GrpcSignalingNode) Apply(i *g.Interface) {
	i.Node = s.Host
}
