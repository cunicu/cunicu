package wg

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"sync"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/conn"
)

var (
	errIncompleteWrite   = errors.New("incomplete write")
	errNoEndpointFound   = errors.New("failed to find endpoint")
	errFailedToParseAddr = errors.New("failed to parse addr from slice")
)

type userPacket struct {
	endpoint *UserEndpoint
	buffer   []byte
}

type UserEndpoint struct {
	conn.StdNetEndpoint

	conn net.Conn
}

type UserBind struct {
	packets chan userPacket

	endpointsLock sync.RWMutex
	endpoints     map[netip.AddrPort]*UserEndpoint

	logger *zap.Logger
}

func NewUserBind() *UserBind {
	return &UserBind{
		endpoints: make(map[netip.AddrPort]*UserEndpoint),
		logger:    zap.L().Named("ice.bind"),
	}
}

// Open puts the Bind into a listening state on a given port and reports the actual
// port that it bound to. Passing zero results in a random selection.
// fns is the set of functions that will be called to receive packets.
func (b *UserBind) Open(port uint16) ([]conn.ReceiveFunc, uint16, error) {
	// b.logger.Debug("Open", zap.Uint16("port", port))

	b.packets = make(chan userPacket)

	return []conn.ReceiveFunc{b.receive}, port, nil
}

// Close closes the Bind listener.
// All fns returned by Open must return net.ErrClosed after a call to Close.
func (b *UserBind) Close() error {
	// b.logger.Debug("Close")

	if b.packets != nil {
		close(b.packets)
	}

	return nil
}

// SetMark sets the mark for each packet sent through this Bind.
// This mark is passed to the kernel as the socket option SO_MARK.
func (b *UserBind) SetMark(mark uint32) error {
	// b.logger.Debug("SetMark", zap.Uint32("mark", mark))

	return nil // Stub
}

// Send writes a packet b to address ep.
func (b *UserBind) Send(buf []byte, ep conn.Endpoint) error {
	uep, ok := ep.(*UserEndpoint)
	if !ok {
		panic("invalid endpoint type")
	}

	// b.logger.Debug("Send",
	// 	zap.Int("len", len(buf)),
	// 	zap.Any("ep", uep),
	// 	zap.String("data", hex.EncodeToString(buf)))

	if n, err := uep.conn.Write(buf); err != nil {
		return fmt.Errorf("failed to write: %w", err)
	} else if n != len(buf) {
		return fmt.Errorf("%w: %d != %d", errIncompleteWrite, n, len(buf))
	}

	return nil
}

// ParseEndpoint creates a new endpoint from a string.
func (b *UserBind) ParseEndpoint(s string) (ep conn.Endpoint, err error) {
	// b.logger.Debug("ParseEndpoint", zap.String("ep", s))

	ap, err := netip.ParseAddrPort(s)
	if err != nil {
		return nil, err
	}

	b.endpointsLock.RLock()
	defer b.endpointsLock.RUnlock()

	ep, ok := b.endpoints[ap]
	if !ok {
		return nil, errNoEndpointFound
	}

	return ep, nil
}

func (b *UserBind) UpdateEndpoint(ep *net.UDPAddr, c net.Conn) (*UserEndpoint, error) {
	// b.logger.Debug("UpdateEndpoint", zap.Any("ep", ep))

	// Remove v4-in-v6 prefix
	epIP := ep.IP
	if epIPv4 := epIP.To4(); epIPv4 != nil {
		epIP = epIPv4
	}

	a, ok := netip.AddrFromSlice(epIP)
	if !ok {
		return nil, errFailedToParseAddr
	}

	ap := netip.AddrPortFrom(a, uint16(ep.Port))

	uEP := &UserEndpoint{
		StdNetEndpoint: conn.StdNetEndpoint(ap),
		conn:           c,
	}

	b.endpointsLock.Lock()
	defer b.endpointsLock.Unlock()

	// TODO: Remove old endpoints
	b.endpoints[ap] = uEP

	return uEP, nil
}

func (b *UserBind) receive(buf []byte) (int, conn.Endpoint, error) {
	pkt, ok := <-b.packets
	if !ok {
		return -1, nil, net.ErrClosed
	}

	n := copy(buf, pkt.buffer)

	// b.logger.Debug("Receive",
	// 	zap.Int("len", n),
	// 	zap.Any("ep", pkt.endpoint),
	// 	zap.String("data", hex.EncodeToString(buf[:n])))

	return n, pkt.endpoint, nil
}

func (b *UserBind) OnData(buf []byte, ep *UserEndpoint) error {
	b.packets <- userPacket{
		endpoint: ep,
		buffer:   buf,
	}

	return nil
}

func (ep *UserEndpoint) String() string {
	return ep.DstToString()
}
