//go:build linux

package e2e

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pion/ice/v2"
	g "github.com/stv0g/gont/pkg"
	nl "github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/internal/wg"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/socket"
)

// Agent is a host running WICE
type Agent struct {
	*g.Host

	Address net.IPNet

	ExtraArgs []interface{}
	Command   *exec.Cmd
	Client    *socket.Client

	WireguardPrivateKey    crypto.Key
	WireguardClient        *wgctrl.Client
	WireguardInterfaceName string
	WireguardListenPort    int

	ID              peer.ID
	ListenAddresses []multiaddr.Multiaddr

	logger zap.Logger
}

var (
	// Singleton for compiled wice executable
	wiceBinary string
)

func NewAgent(m *g.Network, name string, addr net.IPNet, opts ...g.Option) (*Agent, error) {
	h, err := m.AddHost(name, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	a := &Agent{
		Host:            h,
		Address:         addr,
		ListenAddresses: []multiaddr.Multiaddr{},
		ExtraArgs:       []interface{}{},

		WireguardListenPort:    51822,
		WireguardInterfaceName: "wg0",

		logger: *zap.L().Named("agent." + name),
	}

	if err := a.RunFunc(func() error {
		a.WireguardClient, err = wgctrl.New()
		return err
	}); err != nil {
		return nil, fmt.Errorf("failed to create Wireguard client: %w", err)
	}

	if err := a.AddWireguardInterface(); err != nil {
		return nil, fmt.Errorf("failed to create wireguard interface: %w", err)
	}

	return a, nil
}

func NewAgents(n *g.Network, numNodes int, opts ...g.Option) (AgentList, error) {
	al := AgentList{}

	for i := 1; i <= numNodes; i++ {
		addr := net.IPNet{
			IP:   net.IPv4(172, 16, 0, byte(i)),
			Mask: net.IPv4Mask(255, 255, 0, 0),
		}

		node, err := NewAgent(n, fmt.Sprintf("n%d", i), addr, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create node: %w", err)
		}

		al = append(al, node)
	}

	return al, nil
}

func (a *Agent) Start(directArgs ...interface{}) error {
	var err error

	var sockPath = fmt.Sprintf("/var/run/wice.%s.sock", a.Name())
	var logPath = fmt.Sprintf("logs/%s.log", a.Name())

	if err := os.RemoveAll(logPath); err != nil {
		return fmt.Errorf("failed to remove old log file: %w", err)
	}

	args := []interface{}{
		"daemon",
		"--socket", sockPath,
		"--socket-wait",
		"--log-file", logPath,
		"--log-level", "debug",
	}
	args = append(args, directArgs...)
	args = append(args, a.ExtraArgs...)

	if err := os.RemoveAll(sockPath); err != nil {
		log.Fatal(err)
	}

	cmd, err := buildWICE(a.Network())
	if err != nil {
		return fmt.Errorf("failed to build wice: %w", err)
	}

	go func() {
		var out []byte
		if out, a.Command, err = a.Host.Run(cmd, args...); err != nil {
			a.logger.Error("Failed to start", zap.Error(err))
		}

		os.Stdout.Write(out)
	}()

	if a.Client, err = socket.Connect(sockPath); err != nil {
		return fmt.Errorf("failed to connect to to control socket: %w", err)
	}

	return nil
}

func (a *Agent) Stop() error {
	if a.Command == nil || a.Command.Process == nil {
		return nil
	}

	return a.Command.Process.Kill()
}

func (a *Agent) Close() error {
	if a.Client != nil {
		if err := a.Client.Close(); err != nil {
			return fmt.Errorf("failed to close RPC connection: %s", err)
		}
	}

	return a.Stop()
}

func (a *Agent) AddWireguardInterface() error {
	var err error

	a.WireguardInterfaceName = "wg0"
	a.WireguardPrivateKey, err = crypto.GeneratePrivateKey()
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	l := &nl.Wireguard{
		LinkAttrs: nl.NewLinkAttrs(),
	}
	l.LinkAttrs.Name = a.WireguardInterfaceName

	nlh := a.NetlinkHandle()

	if err := nlh.LinkAdd(l); err != nil {
		return fmt.Errorf("failed to create link: %w", err)
	}

	if err := nlh.LinkSetUp(l); err != nil {
		return fmt.Errorf("failed to set link up: %w", err)
	}

	nlAddr := nl.Addr{
		IPNet: &a.Address,
	}

	if err := nlh.AddrAdd(l, &nlAddr); err != nil {
		return fmt.Errorf("failed to assign IP address: %w", err)
	}

	pk := wgtypes.Key(a.WireguardPrivateKey)

	cfg := wgtypes.Config{
		PrivateKey: &pk,
		ListenPort: &a.WireguardListenPort,
	}

	return a.ConfigureWireguardInterface(cfg)
}

func (a *Agent) AddWireguardPeer(peer *Agent) error {
	cfg := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey: wgtypes.Key(peer.WireguardPrivateKey.PublicKey()),
				AllowedIPs: []net.IPNet{
					{
						IP:   peer.Address.IP,
						Mask: net.CIDRMask(32, 32),
					},
				},
			},
		},
	}

	return a.ConfigureWireguardInterface(cfg)
}

func (a *Agent) ConfigureWireguardInterface(cfg wgtypes.Config) error {
	wgCfg := wg.Config{Config: cfg}
	wgCfg.Dump(os.Stdout)

	if err := a.RunFunc(func() error {
		return a.WireguardClient.ConfigureDevice(a.WireguardInterfaceName, cfg)
	}); err != nil {
		return fmt.Errorf("failed to configure Wireguard link: %w", err)
	}

	return nil
}

func (a *Agent) DumpWireguardInterfaces() error {
	return a.RunFunc(func() error {
		devs, err := a.WireguardClient.Devices()
		if err != nil {
			return err
		}

		for _, dev := range devs {
			d := wg.Device(*dev)
			d.DumpEnv(os.Stdout)
		}

		return nil
	})
}

func (a *Agent) WaitReady(p *Agent) error {
	a.Client.WaitForPeerConnectionState(p.WireguardPrivateKey.PublicKey(), ice.ConnectionStateConnected)

	return nil
}

func (a *Agent) PingWireguardPeer(peer *Agent) error {
	if out, _, err := a.Run("ping", "-c", 1, peer.Address.IP); err != nil {
		os.Stdout.Write(out)

		return err
	}

	return nil
}

func (a *Agent) WaitBackendReady() error {
	var err error

	evt := a.Client.WaitForEvent(pb.Event_BACKEND_READY, "", crypto.Key{})

	if be, ok := evt.Event.(*pb.Event_BackendReady); ok {
		a.ID, err = peer.Decode(be.BackendReady.Id)
		if err != nil {
			return fmt.Errorf("failed to decode peer ID: %w", err)
		}

		for _, la := range be.BackendReady.ListenAddresses {
			if ma, err := multiaddr.NewMultiaddr(la); err != nil {
				return fmt.Errorf("failed to decode listen address: %w", err)
			} else {
				a.ListenAddresses = append(a.ListenAddresses, ma)
			}
		}
	} else {
		zap.L().Warn("Missing signaling details")
	}

	return nil
}

func (a *Agent) Dump() {
	a.logger.Info("Details for agent")

	a.DumpWireguardInterfaces()
	a.Run("ip", "addr", "show")
}
