// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package nodes

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	copt "github.com/stv0g/gont/v2/pkg/options/cmd"
	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/wg"
)

var errNoTunnelAddr = errors.New("no WireGuard tunnel address configured")

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

	Agent *Agent

	configLock sync.Mutex
}

func NewWireGuardInterface(name string) (*WireGuardInterface, error) {
	sk, err := crypto.GeneratePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	return &WireGuardInterface{
		Name:                 name,
		SetupKernelInterface: true,
		WriteConfigFile:      false,
		Config: wgtypes.Config{
			PrivateKey: (*wgtypes.Key)(&sk),
		},
	}, nil
}

func (i *WireGuardInterface) String() string {
	return fmt.Sprintf("%s/%s", i.Agent.Name(), i.Name)
}

func (i *WireGuardInterface) Apply(a *Agent) {
	if i.Agent != nil {
		panic("can not assign interface to more than a single agent")
	}

	i.Agent = a

	a.WireGuardInterfaces = append(a.WireGuardInterfaces, i)
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
	wgcpath := i.Agent.Shadowed(wg.ConfigPath)

	fn := filepath.Join(wgcpath, fmt.Sprintf("%s.conf", i.Name))

	f, err := os.OpenFile(fn, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}

	i.Agent.logger.Debug("Writing config file",
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

	nlh := i.Agent.NetlinkHandle()

	if err := nlh.LinkAdd(l); err != nil {
		return fmt.Errorf("failed to create link: %w", err)
	}

	if err := nlh.LinkSetUp(l); err != nil {
		return fmt.Errorf("failed to set link up: %w", err)
	}

	for _, addr := range i.Addresses {
		addr := addr
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

	i.configLock.Lock()
	defer i.configLock.Unlock()

	i.Peers = append(i.Peers, wgtypes.PeerConfig{
		PublicKey:  peer.PrivateKey.PublicKey(),
		AllowedIPs: aIPs,
	})
}

func (i *WireGuardInterface) PingPeer(ctx context.Context, peer *WireGuardInterface) error {
	if len(peer.Addresses) < 1 {
		return errNoTunnelAddr
	}

	out := &bytes.Buffer{}
	if cmd, err := i.Agent.Host.Run("ping", "-v", "-c", 1, "-i", 0.1, "-w", 120, peer.Addresses[0].IP,
		copt.Combined(out),
		copt.EnvVar("LC_ALL", "C"),
		copt.Context{Context: ctx}); err != nil {
		return fmt.Errorf("ping failed with exit code %d: %w\n%s", cmd.ProcessState.ExitCode(), err, out.String())
	}

	return nil
}

func (i *WireGuardInterface) GetConfig() *wg.Config {
	return &wg.Config{
		Config:  i.Config,
		Address: i.Addresses,
	}
}

func (i *WireGuardInterface) WaitConnectionEstablished(ctx context.Context, p *WireGuardInterface) error {
	sk := crypto.Key(*p.PrivateKey)

	return i.Agent.Client.WaitForPeerState(ctx, sk.PublicKey(), daemon.PeerStateConnected)
}

func (i *WireGuardInterface) Configure(cfg wgtypes.Config) error {
	if err := i.Agent.RunFunc(func() error {
		return i.Agent.WireGuardClient.ConfigureDevice(i.Name, cfg)
	}); err != nil {
		return fmt.Errorf("failed to configure WireGuard link: %w", err)
	}

	return nil
}
