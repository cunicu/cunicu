package nodes

import (
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"os"
	"os/exec"
	"time"

	g "github.com/stv0g/gont/pkg"
	"go.uber.org/zap"
	"riasc.eu/wice/internal/test"
)

const portMin = 1<<15 + 1<<14
const portMax = 1<<16 - 1

type GrpcSignalingNode struct {
	*g.Host

	port int

	Command *exec.Cmd

	logger *zap.Logger
}

func NewGrpcSignalingNode(n *g.Network, name string) (SignalingNode, error) {
	h, err := n.AddHost(name)
	if err != nil {
		return nil, err
	}

	t := &GrpcSignalingNode{
		Host:   h,
		port:   rand.Intn(portMax-portMin) + portMin,
		logger: zap.L().Named("signal." + name),
	}

	return t, nil
}

func (s *GrpcSignalingNode) Start(_ ...interface{}) error {
	var err error
	var logPath = fmt.Sprintf("logs/%s.log", s.Name())

	if err := os.RemoveAll(logPath); err != nil {
		return fmt.Errorf("failed to remove old log file: %w", err)
	}

	var args = []interface{}{
		"signal",
		"--log-level", "debug",
		"--log-file", logPath,
		"--log-file", logPath,
		"--listen", fmt.Sprintf(":%d", s.port),
	}

	cmd, err := test.BuildBinary()
	if err != nil {
		return fmt.Errorf("failed to build wice: %w", err)
	}

	go func() {
		var out []byte
		if out, s.Command, err = s.Host.Run(cmd, args...); err != nil {
			s.logger.Error("Failed to start", zap.Error(err))
		}

		os.Stdout.Write(out)
	}()

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
	for tries := 100; !s.isReachable(); tries-- {
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
