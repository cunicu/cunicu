package nodes

import (
	"fmt"
	"net"
	"net/url"
	"os/exec"
	"time"

	g "github.com/stv0g/gont/pkg"
	"go.uber.org/zap"
)

type GrpcSignalingNode struct {
	*g.Host

	port int

	Command *exec.Cmd

	logger *zap.Logger
}

func NewGrpcSignalingNode(n *g.Network, name string, opts ...g.Option) (*GrpcSignalingNode, error) {
	h, err := n.AddHost(name, opts...)
	if err != nil {
		return nil, err
	}

	t := &GrpcSignalingNode{
		Host:   h,
		port:   8080,
		logger: zap.L().Named("node.signal").With(zap.String("node", name)),
	}

	return t, nil
}

func (s *GrpcSignalingNode) Start(_, dir string, extraArgs ...any) error {
	var err error

	logPath := fmt.Sprintf("%s.log", s.Name())

	binary, profileArgs, err := BuildTestBinary(s.Name())
	if err != nil {
		return fmt.Errorf("failed to build: %w", err)
	}

	args := profileArgs
	args = append(args,
		"signal",
		"--log-level", "debug",
		"--log-file", logPath,
		"--listen", fmt.Sprintf(":%d", s.port),
	)
	args = append(args, extraArgs...)

	if _, _, s.Command, err = s.StartWith(binary, nil, dir, args...); err != nil {
		s.logger.Error("Failed to start", zap.Error(err))
	}

	if err := s.WaitReady(); err != nil {
		return fmt.Errorf("failed to start turn server: %w", err)
	}

	return nil
}

func (s *GrpcSignalingNode) Stop() error {
	if s.Command == nil || s.Command.Process == nil {
		return nil
	}

	s.logger.Info("Stopping signaling node")

	return GracefullyTerminate(s.Command)
}

func (s *GrpcSignalingNode) Close() error {
	return s.Stop()
}

func (s *GrpcSignalingNode) URL() *url.URL {
	return &url.URL{
		Scheme:   "grpc",
		Host:     fmt.Sprintf("%s:%d", s.Name(), s.port),
		RawQuery: "insecure=true",
	}
}

func (s *GrpcSignalingNode) isReachable() bool {
	hostPort := fmt.Sprintf("[%s]:%d", net.IPv6loopback, s.port)

	return s.RunFunc(func() error {
		conn, err := net.Dial("tcp6", hostPort)
		if err != nil {
			return err
		}

		return conn.Close()
	}) == nil
}

func (s *GrpcSignalingNode) WaitReady() error {
	for tries := 1000; !s.isReachable(); tries-- {
		if tries == 0 {
			return fmt.Errorf("timed out")
		}

		time.Sleep(20 * time.Millisecond)
	}

	return nil
}

// Options
func (s *GrpcSignalingNode) Apply(i *g.Interface) {
	i.Node = s.Host
}
