// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package nodes

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	g "github.com/stv0g/gont/v2/pkg"
	copt "github.com/stv0g/gont/v2/pkg/options/cmd"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/log"
	rpcproto "github.com/stv0g/cunicu/pkg/proto/rpc"
	"github.com/stv0g/cunicu/pkg/rpc"
)

type AgentOption interface {
	Apply(a *Agent)
}

// Agent is a host running the cunīcu daemon.
//
// Each agent can have one or more WireGuard interfaces configured which are managed
// by a single daemon.
type Agent struct {
	*g.Host

	Command *g.Cmd
	Client  *rpc.Client

	WireGuardClient *wgctrl.Client

	ExtraArgs           []any
	WireGuardInterfaces []*WireGuardInterface

	logFile io.WriteCloser

	logger *log.Logger
}

func NewAgent(m *g.Network, name string, opts ...g.Option) (*Agent, error) {
	h, err := m.AddHost(name, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	a := &Agent{
		Host: h,

		logger: log.Global.Named("node.agent").With(zap.String("node", name)),
	}

	// Apply agent options
	for _, opt := range opts {
		if aopt, ok := opt.(AgentOption); ok {
			aopt.Apply(a)
		}
	}

	// Get wgctrl handle in host netns
	if err := a.RunFunc(func() error {
		a.WireGuardClient, err = wgctrl.New()
		return err
	}); err != nil {
		return nil, fmt.Errorf("failed to create WireGuard client: %w", err)
	}

	return a, nil
}

func (a *Agent) Start(_, dir string, extraArgs ...any) error {
	var err error
	rpcSockPath := fmt.Sprintf("/var/run/cunicu.%s.sock", a.Name())
	logPath := fmt.Sprintf("%s/%s.log", dir, a.Name())

	// Old RPC sockets are also removed by cunīcu.
	// However we also need to do it here to avoid racing
	// against rpc.Connect() further down here
	if err := os.RemoveAll(rpcSockPath); err != nil {
		return fmt.Errorf("failed to remove old socket: %w", err)
	}

	binary, profileArgs, err := BuildBinary(a.Name())
	if err != nil {
		return fmt.Errorf("failed to build: %w", err)
	}

	a.logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	args := profileArgs
	args = append(args,
		"daemon",
		"--rpc-socket", rpcSockPath,
		"--rpc-wait",
		"--log-level", "debug",
		"--sync-hosts=false",
	)
	args = append(args, a.ExtraArgs...)
	args = append(args, extraArgs...)
	args = append(args,
		copt.Combined(a.logFile),
		copt.Dir(dir),
		copt.EnvVar("CUNICU_EXPERIMENTAL", "1"),
		// copt.EnvVar("PION_LOG", "info"),
		copt.EnvVar("GRPC_GO_LOG_SEVERITY_LEVEL", "debug"),
		copt.EnvVar("GRPC_GO_LOG_VERBOSITY_LEVEL", fmt.Sprintf("%d", 99)),
	)

	if a.Command, err = a.Host.Start(binary, args...); err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	if a.Client, err = rpc.Connect(rpcSockPath); err != nil {
		return fmt.Errorf("failed to connect to to control socket: %w", err)
	}

	a.Client.AddEventHandler(a)

	if err := a.Client.Unwait(); err != nil {
		return fmt.Errorf("failed to unwait agent: %w", err)
	}

	return nil
}

func (a *Agent) Stop() error {
	if a.Command == nil || a.Command.Process == nil {
		return nil
	}

	a.logger.Info("Stopping agent node")

	if err := GracefullyTerminate(a.Command); err != nil {
		return fmt.Errorf("failed to terminate: %w", err)
	}

	if err := a.logFile.Close(); err != nil {
		return fmt.Errorf("failed to close log file: %w", err)
	}

	if err := a.Client.Close(); err != nil {
		return fmt.Errorf("failed to close RPC connection: %w", err)
	}

	return nil
}

func (a *Agent) Close() error {
	if a.Client != nil {
		if err := a.Client.Close(); err != nil {
			return fmt.Errorf("failed to close RPC connection: %w", err)
		}
	}

	return a.Stop()
}

func (a *Agent) WaitBackendReady(ctx context.Context) error {
	_, err := a.Client.WaitForEvent(ctx, rpcproto.EventType_BACKEND_READY, "", crypto.Key{})

	return err
}

func (a *Agent) ConfigureWireGuardInterfaces() error {
	for _, i := range a.WireGuardInterfaces {
		if err := i.Create(); err != nil {
			return err
		}
	}

	return nil
}

func (a *Agent) Shadowed(path string) string {
	for _, ed := range a.EmptyDirs {
		if strings.HasPrefix(path, ed) {
			return filepath.Join(a.BasePath, "files", path)
		}
	}

	return path
}

func (a *Agent) OnEvent(e *rpcproto.Event) {
	if e.Type == rpcproto.EventType_PEER_STATE_CHANGED {
		return // be less verbose
	}

	a.logger.Info("New event", zap.Any("event", e))
}
