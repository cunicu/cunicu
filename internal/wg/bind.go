package wg

import (
	"go.uber.org/zap"
	"golang.zx2c4.com/go118/netip"
	"golang.zx2c4.com/wireguard/conn"
)

type IceBind struct {
	// Interface *intf.Interface

	logger *zap.Logger
}

type IceEndpoint struct {
	netip.AddrPort

	// Peer *intf.Peer

	logger *zap.Logger

	name string
}

func (ep *IceEndpoint) String() string {
	return ep.name
}

// clears the source address
func (ep *IceEndpoint) ClearSrc() {
	ep.logger.Debug("ClearSrc()")
}

// returns the local source address (ip:port)
func (ep *IceEndpoint) SrcToString() string {
	ep.logger.Debug("SrcToString()")

	return ep.name + ":src"
}

// returns the destination address (ip:port)
func (ep *IceEndpoint) DstToString() string {
	ep.logger.Debug("DstToString()")

	return ep.name + ":dst"
}

// used for mac2 cookie calculations
func (ep *IceEndpoint) DstToBytes() []byte {
	ep.logger.Debug("DstToBytes()")

	return []byte(ep.name)
}

func (ep *IceEndpoint) DstIP() netip.Addr {
	ep.logger.Debug("DstIP()")

	return ep.Addr()
}

func (ep *IceEndpoint) SrcIP() netip.Addr {
	ep.logger.Debug("SrcIP()")

	return ep.Addr() // TODO: this is wrong
}

func NewIceBind() conn.Bind {
	return &IceBind{
		// Interface: i,
		logger: zap.L().Named("ice.bind"),
	}
}

// Open puts the Bind into a listening state on a given port and reports the actual
// port that it bound to. Passing zero results in a random selection.
// fns is the set of functions that will be called to receive packets.
func (b *IceBind) Open(port uint16) (fns []conn.ReceiveFunc, actualPort uint16, err error) {
	b.logger.Debug("Open()", zap.Uint16("port", port))

	fns = append(fns, b.receive)

	return fns, 0, nil
}

// Close closes the Bind listener.
// All fns returned by Open must return net.ErrClosed after a call to Close.
func (b *IceBind) Close() error {
	b.logger.Debug("Close()")

	return nil
}

// SetMark sets the mark for each packet sent through this Bind.
// This mark is passed to the kernel as the socket option SO_MARK.
func (b *IceBind) SetMark(mark uint32) error {
	b.logger.Debug("SetMark", zap.Uint32("mark", mark))

	return nil // Stub
}

// Send writes a packet b to address ep.
func (b *IceBind) Send(buf []byte, ep conn.Endpoint) error {
	b.logger.Debug("Send()",
		zap.Int("len", len(buf)),
		zap.Any("ep", ep.(*IceEndpoint)),
	)

	return nil
}

// ParseEndpoint creates a new endpoint from a string.
func (b *IceBind) ParseEndpoint(s string) (ep conn.Endpoint, err error) {
	b.logger.Debug("ParseEndpoints()", zap.String("ep", s))

	addrPort, err := netip.ParseAddrPort(s)
	if err != nil {
		return &IceEndpoint{}, err
	}

	return &IceEndpoint{
		AddrPort: addrPort,
		name:     s,
		logger:   b.logger.With(zap.String("ep", s)),
	}, nil
}

func (b *IceBind) receive(buf []byte) (n int, ep conn.Endpoint, err error) {
	b.logger.Debug("Bind receive()")

	buf[0] = 1

	return 1, &IceEndpoint{}, nil
}
