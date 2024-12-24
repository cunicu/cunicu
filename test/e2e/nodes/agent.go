// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package nodes

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	g "cunicu.li/gont/v2/pkg"
	copt "cunicu.li/gont/v2/pkg/options/cmd"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"

	"cunicu.li/cunicu/pkg/crypto"
	"cunicu.li/cunicu/pkg/log"
	rpcproto "cunicu.li/cunicu/pkg/proto/rpc"
	"cunicu.li/cunicu/pkg/rpc"
	"cunicu.li/cunicu/test/e2e/nodes/options"
)

type AgentOption interface {
	Apply(a *Agent)
}

type AgentConfigOption interface {
	Apply(k *koanf.Koanf)
}

// Agent is a host running the cunīcu daemon.
//
// Each agent can have one or more WireGuard interfaces configured which are managed
// by a single daemon.
type Agent struct {
	*g.Host

	Config          *koanf.Koanf
	Command         *g.Cmd
	Client          *rpc.Client
	WireGuardClient *wgctrl.Client

	WireGuardInterfaces []*WireGuardInterface

	logFile io.WriteCloser
	logger  *log.Logger
}

func NewAgent(m *g.Network, name string, opts ...g.Option) (*Agent, error) {
	h, err := m.AddHost(name, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	a := &Agent{
		Host:   h,
		Config: koanf.New("."),

		logger: log.Global.Named("node.agent").With(zap.String("node", name)),
	}

	// Default config
	options.ConfigMap{
		"experimental": true,
		"log.level":    "debug10",
		"rpc.wait":     true,
		"sync_hosts":   false,
	}.Apply(a.Config)

	// Apply agent options
	for _, opt := range opts {
		switch aopt := opt.(type) {
		case AgentOption:
			aopt.Apply(a)
		case AgentConfigOption:
			aopt.Apply(a.Config)
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

func (a *Agent) Start(_, dir string, args ...any) (err error) {
	cfgPath := fmt.Sprintf("%s/%s.yaml", dir, a.Name())
	logPath := fmt.Sprintf("%s/%s.log", dir, a.Name())
	rpcPath := fmt.Sprintf("/run/cunicu.%s.sock", a.Name())

	// Create agent configuration
	cfg := a.Config.Copy()
	cfg.Set("rpc.socket", rpcPath) //nolint:errcheck

	extraArgs := []any{}
	for _, arg := range args {
		if aopt, ok := arg.(AgentConfigOption); ok {
			aopt.Apply(cfg)
		} else {
			extraArgs = append(extraArgs, arg)
		}
	}

	// Old RPC sockets are also removed by cunīcu.
	// However we also need to do it here to avoid racing
	// against rpc.Connect() further down here
	if err := os.RemoveAll(rpcPath); err != nil {
		return fmt.Errorf("failed to remove old socket: %w", err)
	}

	binary, startArgs, err := BuildBinary(a.Name())
	if err != nil {
		return fmt.Errorf("failed to build: %w", err)
	}

	a.logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	cfgRaw, err := cfg.Marshal(yaml.Parser())
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cfgPath, cfgRaw, 0o644); err != nil { //nolint:gosec
		return fmt.Errorf("failed to write config file: %w", err)
	}

	startArgs = append(startArgs,
		copt.Combined(a.logFile),
		copt.EnvVar("CUNICU_CONFIG_ALLOW_INSECURE", "true"),
		copt.Dir(dir),
		"daemon", "--config", cfgPath,
	)
	startArgs = append(startArgs, extraArgs...)

	if a.Command, err = a.Host.Start(binary, startArgs...); err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	if a.Client, err = rpc.Connect(rpcPath); err != nil {
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
