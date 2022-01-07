//go:build linux

package test

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"

	g "github.com/stv0g/gont/pkg"
	nl "github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/socket"
)

type Node struct {
	*g.Host

	PrivateKey crypto.Key
	Address    net.IPNet

	InterfaceName string
	ListenPort    int

	ExtraArgs []interface{}
	Command   *exec.Cmd
	Client    *socket.Client

	WireguardClient *wgctrl.Client

	Backend *SignalingNode
}

func NewNode(m *g.Network, name string, backend *SignalingNode, addr net.IPNet, opts ...g.Option) (*Node, error) {
	h, err := m.AddHost(name, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	n := &Node{
		Host:          h,
		Backend:       backend,
		Address:       addr,
		InterfaceName: "wg0",
		ListenPort:    51822,
	}

	if n.PrivateKey, err = crypto.GeneratePrivateKey(); err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	if err := n.RunFunc(func() error {
		n.WireguardClient, err = wgctrl.New()
		return err
	}); err != nil {
		return nil, fmt.Errorf("failed to create Wireguard client: %w", err)
	}

	if err := n.SetupInterface(); err != nil {
		return nil, fmt.Errorf("failed to setup interface: %w", err)
	}

	return n, nil
}

func (n *Node) Start(directArgs ...interface{}) error {
	var err error
	var stdout, stderr io.Reader

	var sockPath = fmt.Sprintf("/var/run/wice.%s.sock", n.Name())
	var logPath = fmt.Sprintf("logs/%s.log", n.Name())

	u, err := n.Backend.URL()
	if err != nil {
		return err
	}

	args := []interface{}{
		"daemon",
		"--backend", u.String(),
		"--socket", sockPath,
		"--log-level", "debug"}
	args = append(args, directArgs...)
	args = append(args, n.ExtraArgs...)

	if err := os.RemoveAll(sockPath); err != nil {
		log.Fatal(err)
	}

	if stdout, stderr, n.Command, err = n.StartGo("../cmd/wice", args...); err != nil {
		return err
	}

	if _, err = FileWriter(logPath, stdout, stderr); err != nil {
		return err
	}

	if n.Client, err = socket.Connect(sockPath); err != nil {
		return err
	}

	return nil
}

func (n *Node) Stop() error {
	if n.Command == nil || n.Command.Process == nil {
		return nil
	}

	return n.Command.Process.Kill()
}

func (n *Node) Close() error {
	return n.Stop()
}

func (n *Node) SetupInterface() error {
	nlh := n.NetlinkHandle()

	l := &nl.Wireguard{
		LinkAttrs: nl.NewLinkAttrs(),
	}
	l.LinkAttrs.Name = n.InterfaceName

	if err := nlh.LinkAdd(l); err != nil {
		return fmt.Errorf("failed to create Wireguard interface: %w", err)
	}

	if err := nlh.LinkSetUp(l); err != nil {
		return fmt.Errorf("failed to set Wireguard interface up: %w", err)
	}

	nlAddr := nl.Addr{
		IPNet: &n.Address,
	}

	if err := nlh.AddrAdd(l, &nlAddr); err != nil {
		return fmt.Errorf("failed to assign IP address: %w", err)
	}

	var privKey = wgtypes.Key(n.PrivateKey)

	cfg := wgtypes.Config{
		PrivateKey: &privKey,
		ListenPort: &n.ListenPort,
	}

	return n.ConfigureInterface(cfg)
}

func (n *Node) AddPeer(peer *Node) error {
	cfg := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey: wgtypes.Key(peer.PrivateKey.PublicKey()),
				AllowedIPs: []net.IPNet{
					peer.Address,
				},
			},
		},
	}

	return n.ConfigureInterface(cfg)
}

func (n *Node) ConfigureInterface(cfg wgtypes.Config) error {
	if err := n.RunFunc(func() error {
		return n.WireguardClient.ConfigureDevice(n.InterfaceName, cfg)
	}); err != nil {
		return fmt.Errorf("failed to configure Wireguard link: %w", err)
	}

	return nil
}

func (n *Node) WaitReady(peer *Node) error {
	n.Client.WaitPeerHandshake(peer.PrivateKey.PublicKey())

	return nil
}

func (n *Node) PingPeer(peer *Node) error {
	_, _, err := n.Run("ping", "-c", 1, peer.Address.IP)

	return err
}
