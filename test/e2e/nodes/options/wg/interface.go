package wg

import (
	"fmt"
	"net"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/test/e2e/nodes"
	g "github.com/stv0g/gont/pkg"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type PrivateKey crypto.Key

func (pk PrivateKey) Apply(i *nodes.WireGuardInterface) {
	pkp := wgtypes.Key(pk)
	i.PrivateKey = &pkp
}

type ListenPort int

func (lp ListenPort) Apply(i *nodes.WireGuardInterface) {
	lpp := int(lp)
	i.ListenPort = &lpp
}

type Address net.IPNet

func (addr Address) Apply(i *nodes.WireGuardInterface) {
	i.Addresses = append(i.Addresses, net.IPNet(addr))
}

func AddressIPv4(a, b, c, d byte, m int) Address {
	return Address{
		IP:   net.IPv4(a, b, c, d),
		Mask: net.CIDRMask(m, 32),
	}
}

func AddressIP(fmts string, args ...any) Address {
	str := fmt.Sprintf(fmts, args...)

	ip, n, err := net.ParseCIDR(str)
	if err != nil {
		panic(fmt.Errorf("failed to parse CIDR: %w", err))
	}

	return Address{
		IP:   ip,
		Mask: n.Mask,
	}
}

type WriteConfigFile bool

func (wcf WriteConfigFile) Apply(i *nodes.WireGuardInterface) {
	i.WriteConfigFile = bool(wcf)
}

type SetupKernelInterface bool

func (ski SetupKernelInterface) Apply(i *nodes.WireGuardInterface) {
	i.SetupKernelInterface = bool(ski)
}

type PeerSelector nodes.WireGuardPeerSelectorFunc

var FullMeshPeers PeerSelector = func(i, j *nodes.WireGuardInterface) bool { return true }
var NoPeers PeerSelector = func(i, j *nodes.WireGuardInterface) bool { return false }

func (ps PeerSelector) Apply(i *nodes.WireGuardInterface) {
	i.PeerSelector = nodes.WireGuardPeerSelectorFunc(ps)
}

func Interface(name string, opts ...g.Option) *nodes.WireGuardInterface {
	i, err := nodes.NewWireGuardInterface(name)
	if err != nil {
		panic(fmt.Errorf("failed to create WireGuard interface: %w", err))
	}

	for _, o := range opts {
		switch opt := o.(type) {
		case nodes.WireGuardInterfaceOption:
			opt.Apply(i)
		}
	}

	return i
}
