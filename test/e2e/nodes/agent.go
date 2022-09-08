package nodes

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	g "github.com/stv0g/gont/pkg"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/rpc"
	"github.com/stv0g/cunicu/pkg/wg"

	rpcproto "github.com/stv0g/cunicu/pkg/proto/rpc"
)

type AgentOption interface {
	Apply(a *Agent)
}

// Agent is a host running the ɯice daemon.
//
// Each agent can have one or more WireGuard interfaces configured which are managed
// by a single daemon.
type Agent struct {
	*g.Host

	Command *exec.Cmd
	Client  *rpc.Client

	WireGuardClient *wgctrl.Client

	ExtraArgs           []any
	WireGuardInterfaces []*WireGuardInterface

	// Path of a wg-quick(8) configuration file describing the interface rather than a kernel device
	// Will only be created if non-empty
	WireGuardConfigPath string

	logFile io.WriteCloser

	logger *zap.Logger
}

func NewAgent(m *g.Network, name string, opts ...g.Option) (*Agent, error) {
	h, err := m.AddHost(name, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	a := &Agent{
		Host: h,

		WireGuardInterfaces: []*WireGuardInterface{},
		WireGuardConfigPath: wg.ConfigPath,
		ExtraArgs:           []any{},

		logger: zap.L().Named("node.agent").With(zap.String("node", name)),
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
	var stdout, stderr io.Reader
	var rpcSockPath = fmt.Sprintf("/var/run/wice.%s.sock", a.Name())
	var logPath = fmt.Sprintf("%s/%s.log", dir, a.Name())

	// Old RPC sockets are also removed by wice.
	// However we also need to do it here to avoid racing
	// against rpc.Connect() further down here
	if err := os.RemoveAll(rpcSockPath); err != nil {
		return fmt.Errorf("failed to remove old socket: %w", err)
	}

	binary, profileArgs, err := BuildTestBinary(a.Name())
	if err != nil {
		return fmt.Errorf("failed to build: %w", err)
	}

	args := profileArgs
	args = append(args,
		"daemon",
		"--rpc-socket", rpcSockPath,
		"--rpc-wait",
		"--log-level", "debug",
		"--host-sync=false",
		"--config-path", a.WireGuardConfigPath,
	)
	args = append(args, a.ExtraArgs...)
	args = append(args, extraArgs...)

	env := []string{
		// "PION_LOG=debug",
		fmt.Sprintf("GORACE=log_path=%s-race.log", a.Name()),
	}

	if stdout, stderr, a.Command, err = a.StartWith(binary, env, dir, args...); err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	multi := io.MultiReader(stdout, stderr)

	//#nosec G304 -- Test code is not controllable by attackers
	//#nosec G302 -- Log file should be readable by user
	a.logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	go io.Copy(a.logFile, multi)

	if a.Client, err = rpc.Connect(rpcSockPath); err != nil {
		return fmt.Errorf("failed to connect to to control socket: %w", err)
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

	return nil
}

func (a *Agent) Close() error {
	if a.Client != nil {
		if err := a.Client.Close(); err != nil {
			return fmt.Errorf("failed to close RPC connection: %s", err)
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
