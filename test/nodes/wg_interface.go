//go:build linux

package nodes

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/pion/ice/v2"
	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/wg"
)

type WireGuardInterfaceOption interface {
	Apply(i *WireGuardInterface)
}

type WireGuardInterface struct {
	wgtypes.Config

	// Name of the WireGuard interface
	Name string

	// List of addresses which will be assigned to the interface
	Addresses []net.IPNet

	WriteConfigFile      bool
	SetupKernelInterface bool

	agent *Agent
}

func (i *WireGuardInterface) Apply(a *Agent) {
	if i.agent != nil {
		panic(fmt.Errorf("can not assign interface to more than a single agent"))
	}

	i.agent = a

	a.WireGuardInterfaces = append(a.WireGuardInterfaces, i)
}

func NewWireGuardInterface(name string) *WireGuardInterface {
	lp := wg.DefaultPort

	return &WireGuardInterface{
		Name:                 name,
		SetupKernelInterface: true,
		WriteConfigFile:      false,
		Config: wgtypes.Config{
			ListenPort: &lp,
		},
	}
}

func (i *WireGuardInterface) Create() error {
	// Generate private key if not provided
	if i.PrivateKey == nil {
		sk := crypto.PrivateKeyFromStrings(i.agent.Name(), i.Name)
		i.PrivateKey = (*wgtypes.Key)(&sk)
	}

	if i.WriteConfigFile {
		if err := i.WriteConfig(); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}
	}

	if i.SetupKernelInterface {
		if err := i.SetupKernel(); err != nil {
			return fmt.Errorf("failed to setup kernel interface: %w", err)
		}
	}

	return nil
}

func (i *WireGuardInterface) WriteConfig() error {
	wgcpath := i.agent.Shadowed(i.agent.WireGuardConfigPath)

	fn := filepath.Join(wgcpath, fmt.Sprintf("%s.conf", i.Name))
	f, err := os.OpenFile(fn, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}

	i.agent.logger.Debug("Writing config file",
		zap.String("intf", i.Name),
		zap.String("path", fn))

	if err := i.GetConfig().Dump(f); err != nil {
		return fmt.Errorf("failed to write config to file: %w", err)
	}

	return nil
}

func (i *WireGuardInterface) SetupKernel() error {
	l := &netlink.Wireguard{
		LinkAttrs: netlink.NewLinkAttrs(),
	}
	l.LinkAttrs.Name = i.Name

	nlh := i.agent.NetlinkHandle()

	if err := nlh.LinkAdd(l); err != nil {
		return fmt.Errorf("failed to create link: %w", err)
	}

	if err := nlh.LinkSetUp(l); err != nil {
		return fmt.Errorf("failed to set link up: %w", err)
	}

	for _, addr := range i.Addresses {
		nlAddr := netlink.Addr{
			IPNet: &addr,
		}

		if err := nlh.AddrAdd(l, &nlAddr); err != nil {
			return fmt.Errorf("failed to assign IP address: %w", err)
		}
	}

	return i.Configure(i.Config)
}

func (i *WireGuardInterface) AddPeer(peer *WireGuardInterface) error {
	aIPs := []net.IPNet{}
	for _, addr := range peer.Addresses {
		var mask net.IPMask
		if addr.IP.To4() == nil { // is IPv6
			mask = net.CIDRMask(128, 128)
		} else {
			mask = net.CIDRMask(32, 32)
		}

		aIPs = append(aIPs, net.IPNet{
			IP:   addr.IP,
			Mask: mask,
		})
	}

	cfg := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey:  wgtypes.Key(peer.PrivateKey.PublicKey()),
				AllowedIPs: aIPs,
			},
		},
	}

	return i.Configure(cfg)
}

func (i *WireGuardInterface) PingPeer(peer *WireGuardInterface) error {
	os.Setenv("LC_ALL", "C") // fix issues with parsing of -W and -i options

	if len(peer.Addresses) < 1 {
		return fmt.Errorf("no WireGuard tunnel address configured")
	}

	if out, _, err := i.agent.Run("ping", "-c", 1, "-w", 15, "-i", 0.2, peer.Addresses[0].IP); err != nil {
		os.Stdout.Write(out)
		os.Stdout.Sync()

		return err
	}

	return nil
}

func (i *WireGuardInterface) GetConfig() *wg.Config {
	return &wg.Config{
		Config:  i.Config,
		Address: i.Addresses,
	}
}

func (i *WireGuardInterface) WaitConnectionReady(p *WireGuardInterface) error {
	sk := crypto.Key(*p.PrivateKey)
	i.agent.Client.WaitForPeerConnectionState(sk.PublicKey(), ice.ConnectionStateConnected)

	return nil
}

func (i *WireGuardInterface) Configure(cfg wgtypes.Config) error {
	if err := i.agent.RunFunc(func() error {
		return i.agent.WireGuardClient.ConfigureDevice(i.Name, cfg)
	}); err != nil {
		return fmt.Errorf("failed to configure WireGuard link: %w", err)
	}

	return nil
}
