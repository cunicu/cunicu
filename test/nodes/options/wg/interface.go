package options

import (
	"net"

	g "github.com/stv0g/gont/pkg"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/test/nodes"
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

func AddressIP(str string) Address {
	ip, n, _ := net.ParseCIDR(str)

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

func Interface(name string, opts ...g.Option) *nodes.WireGuardInterface {
	i := nodes.NewWireGuardInterface(name)

	for _, o := range opts {
		switch opt := o.(type) {
		case nodes.WireGuardInterfaceOption:
			opt.Apply(i)
		}
	}

	return i
}
