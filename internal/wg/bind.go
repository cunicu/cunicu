package wg

import (
	"net"

	log "github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/conn"
)

type IceBind struct {
	// Interface *intf.Interface
}

type IceEndpoint struct {
	net.UDPAddr

	// Peer *intf.Peer

	String string
}

// clears the source address
func (ep *IceEndpoint) ClearSrc() {
	log.Debugf("EP %s ClearSrc()", ep.String)
}

// returns the local source address (ip:port)
func (ep *IceEndpoint) SrcToString() string {
	log.Debugf("EP %s SrcToString()", ep.String)

	return ep.String + ":src"
}

// returns the destination address (ip:port)
func (ep *IceEndpoint) DstToString() string {
	log.Debugf("EP %s DstToString()", ep.String)

	return ep.String + ":dst"
}

// used for mac2 cookie calculations
func (ep *IceEndpoint) DstToBytes() []byte {
	log.Debugf("EP %s DstToBytes()", ep.String)

	return []byte(ep.String)
}

func (ep *IceEndpoint) DstIP() net.IP {
	log.Debugf("EP %s DstIP()", ep.String)

	return ep.IP
}

func (ep *IceEndpoint) SrcIP() net.IP {
	log.Debugf("EP %s SrcIP()", ep.String)

	return ep.IP
}

func NewIceBind() conn.Bind {
	return &IceBind{
		// Interface: i,
	}
}

// Open puts the Bind into a listening state on a given port and reports the actual
// port that it bound to. Passing zero results in a random selection.
// fns is the set of functions that will be called to receive packets.
func (b *IceBind) Open(port uint16) (fns []conn.ReceiveFunc, actualPort uint16, err error) {
	log.Debugf("Bind Open(port=%d)", port)

	fns = append(fns, b.receive)

	return fns, 0, nil
}

// Close closes the Bind listener.
// All fns returned by Open must return net.ErrClosed after a call to Close.
func (b *IceBind) Close() error {
	log.Debug("Bind Close()")

	return nil
}

// SetMark sets the mark for each packet sent through this Bind.
// This mark is passed to the kernel as the socket option SO_MARK.
func (b *IceBind) SetMark(mark uint32) error {
	log.Debugf("Bind SetMark(mark=%d)", mark)

	return nil // Stub
}

// Send writes a packet b to address ep.
func (b *IceBind) Send(buf []byte, ep conn.Endpoint) error {
	log.Debugf("Bind Send(len=%d, ep=%s)", len(buf), ep.(*IceEndpoint).String)

	return nil
}

// ParseEndpoint creates a new endpoint from a string.
func (b *IceBind) ParseEndpoint(s string) (ep conn.Endpoint, err error) {
	log.Debugf("Bind ParseEndpoints(%s)", s)

	addr, err := net.ResolveUDPAddr("udp", s)
	if err != nil {
		return &IceEndpoint{}, err
	}

	return &IceEndpoint{
		UDPAddr: *addr,
		String:  s,
	}, nil
}

func (b *IceBind) receive(buf []byte) (n int, ep conn.Endpoint, err error) {
	log.Debug("Bind receive()")

	buf[0] = 1

	return 1, &IceEndpoint{}, nil
}
