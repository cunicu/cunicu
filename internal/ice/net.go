package ice

import (
	"fmt"
	"net"
	"os"

	"github.com/pion/transport/vnet"
)

// Net represents a local network stack euivalent to a set of layers from NIC
// up to the transport (UDP / TCP) layer.
type Net struct {
	ifs []*vnet.Interface
}

type UDPPacketConn struct {
	net.PacketConn
}

func (c *UDPPacketConn) Read(b []byte) (int, error) {
	return 0, nil
}

func (c *UDPPacketConn) RemoteAddr() net.Addr {
	return &net.IPAddr{}
}

func (c *UDPPacketConn) Write(b []byte) (int, error) {
	return 0, nil
}

func NewNet() *Net {
	ifs := []*vnet.Interface{}
	if orgIfs, err := net.Interfaces(); err == nil {
		for _, orgIfc := range orgIfs {
			ifc := vnet.NewInterface(orgIfc)
			if addrs, err := orgIfc.Addrs(); err == nil {
				for _, addr := range addrs {
					ifc.AddAddr(addr)
				}
			}

			ifs = append(ifs, ifc)
		}
	}

	return &Net{ifs: ifs}
}

// Interfaces returns a list of the system's network interfaces.
func (n *Net) Interfaces() ([]*vnet.Interface, error) {
	return n.ifs, nil
}

// InterfaceByName returns the interface specified by name.
func (n *Net) InterfaceByName(name string) (*vnet.Interface, error) {
	for _, ifc := range n.ifs {
		if ifc.Name == name {
			return ifc, nil
		}
	}

	return nil, fmt.Errorf("interface %s: %w", name, os.ErrNotExist)
}

// ListenPacket announces on the local network address.
func (n *Net) ListenPacket(network string, address string) (net.PacketConn, error) {
	return net.ListenPacket(network, address)
}

// ListenUDP acts like ListenPacket for UDP networks.
func (n *Net) ListenUDP(network string, locAddr *net.UDPAddr) (vnet.UDPPacketConn, error) {
	return net.ListenUDP(network, locAddr)
}

// Dial connects to the address on the named network.
func (n *Net) Dial(network, address string) (net.Conn, error) {
	return net.Dial(network, address)
}

// CreateDialer creates an instance of vnet.Dialer
func (n *Net) CreateDialer(dialer *net.Dialer) Dialer {
	return &vDialer{
		dialer: dialer,
	}
}

// DialUDP acts like Dial for UDP networks.
func (n *Net) DialUDP(network string, laddr, raddr *net.UDPAddr) (vnet.UDPPacketConn, error) {
	conn, err := net.DialUDP(network, laddr, raddr)
	if err != nil {
		return nil, err
	}

	return &UDPPacketConn{
		PacketConn: conn,
	}, nil
}

// ResolveUDPAddr returns an address of UDP end point.
func (n *Net) ResolveUDPAddr(network, address string) (*net.UDPAddr, error) {
	return net.ResolveUDPAddr(network, address)
}

// IsVirtual tests if the virtual network is enabled.
func (n *Net) IsVirtual() bool {
	return false
}

// Dialer is identical to net.Dialer excepts that its methods
// (Dial, DialContext) are overridden to use virtual network.
// Use vnet.CreateDialer() to create an instance of this Dialer.
type Dialer interface {
	Dial(network, address string) (net.Conn, error)
}

type vDialer struct {
	dialer *net.Dialer
}

func (d *vDialer) Dial(network, address string) (net.Conn, error) {
	return d.dialer.Dial(network, address)
}
