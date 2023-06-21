// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package nodes

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"time"

	g "github.com/stv0g/gont/v2/pkg"
	copt "github.com/stv0g/gont/v2/pkg/options/cmd"
	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/log"
)

type GrpcSignalingNode struct {
	*g.Host

	port int

	Command *g.Cmd

	logFile io.WriteCloser
	logger  *log.Logger
}

func NewGrpcSignalingNode(n *g.Network, name string, opts ...g.Option) (*GrpcSignalingNode, error) {
	h, err := n.AddHost(name, opts...)
	if err != nil {
		return nil, err
	}

	t := &GrpcSignalingNode{
		Host:   h,
		port:   8080,
		logger: log.Global.Named("node.signal").With(zap.String("node", name)),
	}

	return t, nil
}

func (s *GrpcSignalingNode) Start(_, dir string, extraArgs ...any) error {
	logPath := fmt.Sprintf("%s/%s.log", dir, s.Name())

	binary, profileArgs, err := BuildBinary(s.Name())
	if err != nil {
		return fmt.Errorf("failed to build: %w", err)
	}

	s.logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	args := profileArgs
	args = append(args,
		"signal",
		"--log-level", "debug",
		"--listen", fmt.Sprintf(":%d", s.port),
	)
	args = append(args, extraArgs...)
	args = append(args,
		copt.Dir(dir),
		copt.Combined(s.logFile),
		// copt.EnvVar("GOMAXPROCS", "10"),
		copt.EnvVar("GRPC_GO_LOG_SEVERITY_LEVEL", "debug"),
		copt.EnvVar("GRPC_GO_LOG_VERBOSITY_LEVEL", "99"),
		copt.EnvVar("GORACE", fmt.Sprintf("log_path=%s-race.log", s.Name())),
	)

	if s.Command, err = s.Host.Start(binary, args...); err != nil {
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

	if err := GracefullyTerminate(s.Command); err != nil {
		return err
	}

	if err := s.logFile.Close(); err != nil {
		return fmt.Errorf("failed to close log file: %w", err)
	}

	return nil
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
			return errTimeout
		}

		time.Sleep(20 * time.Millisecond)
	}

	return nil
}

// Options
func (s *GrpcSignalingNode) Apply(i *g.Interface) {
	i.Node = s.Host
}
