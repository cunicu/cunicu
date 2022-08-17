//go:build linux

package nodes

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/pion/ice/v2"
	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/wg"
)

type WireGuardPeerSelectorFunc func(i, j *WireGuardInterface) bool

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
	PeerSelector         WireGuardPeerSelectorFunc

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
	sk, _ := crypto.GeneratePrivateKey()

	return &WireGuardInterface{
		Name:                 name,
		SetupKernelInterface: true,
		WriteConfigFile:      false,
		Config: wgtypes.Config{
			ListenPort: &lp,
			PrivateKey: (*wgtypes.Key)(&sk),
		},
	}
}

func (i *WireGuardInterface) Create() error {
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

func (i *WireGuardInterface) AddPeer(peer *WireGuardInterface) {
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

	i.Peers = append(i.Peers, wgtypes.PeerConfig{
		PublicKey:  wgtypes.Key(peer.PrivateKey.PublicKey()),
		AllowedIPs: aIPs,
	})
}

func (i *WireGuardInterface) PingPeer(ctx context.Context, peer *WireGuardInterface) error {
	env := []string{"LC_ALL=C"} // fix issues with parsing of -W and -i options

	if len(peer.Addresses) < 1 {
		return fmt.Errorf("no WireGuard tunnel address configured")
	}

	stdout, stderr, cmd, err := i.agent.Host.StartWith("ping", env, "", "-c", 1, "-i", 0.2, "-w", time.Hour.Seconds(), i.Addresses[0].IP)
	if err != nil {
		return fmt.Errorf("failed to start ping process: %w", err)
	}

	out := []byte{}
	errs := make(chan error)
	go func() {
		combined := io.MultiReader(stdout, stderr)
		out, _ = io.ReadAll(combined)

		errs <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		cmd.Process.Kill()
		return ctx.Err()
	case err := <-errs:
		if err != nil {
			return fmt.Errorf("ping failed with exit code %d: %w\n%s", cmd.ProcessState.ExitCode(), err, out)
		} else {
			i.agent.logger.Info("Pinged successfully",
				zap.String("intf", i.Name),
				zap.String("peer", peer.agent.Name()),
				zap.String("peer_intf", peer.Name))

			return nil
		}
	}
}

func (i *WireGuardInterface) GetConfig() *wg.Config {
	return &wg.Config{
		Config:  i.Config,
		Address: i.Addresses,
	}
}

func (i *WireGuardInterface) WaitConnectionReady(ctx context.Context, p *WireGuardInterface) error {
	sk := crypto.Key(*p.PrivateKey)

	return i.agent.Client.WaitForPeerConnectionState(ctx, sk.PublicKey(), ice.ConnectionStateConnected)
}

func (i *WireGuardInterface) Configure(cfg wgtypes.Config) error {
	if err := i.agent.RunFunc(func() error {
		return i.agent.WireGuardClient.ConfigureDevice(i.Name, cfg)
	}); err != nil {
		return fmt.Errorf("failed to configure WireGuard link: %w", err)
	}

	return nil
}
