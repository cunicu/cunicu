package ice

// Based on https://github.com/pion/ice/blob/v2.1.8/udp_mux.go

import (
	"io"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/pion/stun"
	"go.uber.org/zap"

	netx "riasc.eu/wice/internal/net"
)

const (
	receiveMTU  = 8192
	maxAddrSize = 512
)

// UDPMuxDefault is an implementation of the interface
type FilteredUDPMux struct {
	conn *netx.FilteredUDPConn

	closedChan chan struct{}
	closeOnce  sync.Once

	// conns is a map of all udpMuxedConn indexed by ufrag|network|candidateType
	conns map[string]*udpMuxedConn

	addressMapMu sync.RWMutex
	addressMap   map[string]*udpMuxedConn

	// buffer pool to recycle buffers for net.UDPAddr encodes/decodes
	pool *sync.Pool

	mu sync.Mutex

	logger *zap.Logger
}

// NewUDPMuxDefault creates an implementation of UDPMux
func NewFilteredUDPMux(conn *netx.FilteredUDPConn) *FilteredUDPMux {
	m := &FilteredUDPMux{
		addressMap: map[string]*udpMuxedConn{},
		conns:      make(map[string]*udpMuxedConn),
		closedChan: make(chan struct{}, 1),
		pool: &sync.Pool{
			New: func() interface{} {
				// big enough buffer to fit both packet and address
				return newBufferHolder(receiveMTU + maxAddrSize)
			},
		},
		logger: zap.L().Named("ice.mux"),
	}

	go m.connWorker()

	return m
}

// LocalAddr returns the listening address of this UDPMuxDefault
func (m *FilteredUDPMux) LocalAddr() net.Addr {
	return m.conn.LocalAddr()
}

// GetConn returns a PacketConn given the connection's ufrag and network
// creates the connection if an existing one can't be found
func (m *FilteredUDPMux) GetConn(ufrag string) (net.PacketConn, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.IsClosed() {
		return nil, io.ErrClosedPipe
	}

	if c, ok := m.conns[ufrag]; ok {
		return c, nil
	}

	c := m.createMuxedConn(ufrag)
	go func() {
		<-c.CloseChannel()
		m.removeConn(ufrag)
	}()
	m.conns[ufrag] = c
	return c, nil
}

// RemoveConnByUfrag stops and removes the muxed packet connection
func (m *FilteredUDPMux) RemoveConnByUfrag(ufrag string) {
	m.mu.Lock()
	removedConns := make([]*udpMuxedConn, 0)
	for key := range m.conns {
		if key != ufrag {
			continue
		}

		c := m.conns[key]
		delete(m.conns, key)
		if c != nil {
			removedConns = append(removedConns, c)
		}
	}
	// keep lock section small to avoid deadlock with conn lock
	m.mu.Unlock()

	m.addressMapMu.Lock()
	defer m.addressMapMu.Unlock()

	for _, c := range removedConns {
		addresses := c.getAddresses()
		for _, addr := range addresses {
			delete(m.addressMap, addr)
		}
	}
}

// IsClosed returns true if the mux had been closed
func (m *FilteredUDPMux) IsClosed() bool {
	select {
	case <-m.closedChan:
		return true
	default:
		return false
	}
}

// Close the mux, no further connections could be created
func (m *FilteredUDPMux) Close() error {
	m.logger.Info("Closing mux")

	var err error
	m.closeOnce.Do(func() {
		m.mu.Lock()
		defer m.mu.Unlock()

		for _, c := range m.conns {
			_ = c.Close()
		}
		m.conns = make(map[string]*udpMuxedConn)
		close(m.closedChan)
	})
	return err
}

func (m *FilteredUDPMux) removeConn(key string) {
	m.mu.Lock()
	c := m.conns[key]
	delete(m.conns, key)
	// keep lock section small to avoid deadlock with conn lock
	m.mu.Unlock()

	if c == nil {
		return
	}

	m.addressMapMu.Lock()
	defer m.addressMapMu.Unlock()

	addresses := c.getAddresses()
	for _, addr := range addresses {
		delete(m.addressMap, addr)
	}
}

func (m *FilteredUDPMux) writeTo(buf []byte, raddr net.Addr) (n int, err error) {
	return m.conn.WriteTo(buf, raddr)
}

func (m *FilteredUDPMux) registerConnForAddress(conn *udpMuxedConn, addr string) {
	if m.IsClosed() {
		return
	}

	m.addressMapMu.Lock()
	defer m.addressMapMu.Unlock()

	existing, ok := m.addressMap[addr]
	if ok {
		existing.removeAddress(addr)
	}
	m.addressMap[addr] = conn

	m.logger.Sugar().Debugf("Registered %s for %s", addr, conn.params.Key)
}

func (m *FilteredUDPMux) createMuxedConn(key string) *udpMuxedConn {
	c := newUDPMuxedConn(&udpMuxedConnParams{
		Mux:       m,
		Key:       key,
		AddrPool:  m.pool,
		LocalAddr: m.LocalAddr(),
	})
	return c
}

func (m *FilteredUDPMux) connWorker() {
	defer func() {
		_ = m.Close()
	}()

	buf := make([]byte, receiveMTU)
	for {
		n, addr, err := m.conn.ReadFrom(buf)
		if m.IsClosed() {
			return
		} else if err != nil {
			if os.IsTimeout(err) {
				continue
			} else if err != io.EOF {
				m.logger.Error("Could not read UDP packet", zap.Error(err))
			}

			return
		}

		udpAddr, ok := addr.(*net.UDPAddr)
		if !ok {
			m.logger.Error("Underlying PacketConn did not return a UDPAddr")
			return
		}

		// If we have already seen this address dispatch to the appropriate destination
		m.addressMapMu.Lock()
		destinationConn := m.addressMap[addr.String()]
		m.addressMapMu.Unlock()

		// If we haven't seen this address before but is a STUN packet lookup by ufrag
		if destinationConn == nil && stun.IsMessage(buf[:n]) {
			msg := &stun.Message{
				Raw: append([]byte{}, buf[:n]...),
			}

			if err = msg.Decode(); err != nil {
				m.logger.Warn("Failed to handle decode ICE", zap.Any("address", addr), zap.Error(err))
				continue
			}

			attr, stunAttrErr := msg.Get(stun.AttrUsername)
			if stunAttrErr != nil {
				m.logger.Warn("No Username attribute in STUN message", zap.Any("address", addr))
				continue
			}

			ufrag := strings.Split(string(attr), ":")[0]

			m.mu.Lock()
			destinationConn = m.conns[ufrag]
			m.mu.Unlock()
		}

		if destinationConn == nil {
			m.logger.Sugar().Debug("Dropping packet from %s", udpAddr.String(), zap.Any("address", addr))
			continue
		}

		if err = destinationConn.writePacket(buf[:n], udpAddr); err != nil {
			m.logger.Error("Could not write packet: %v", zap.Error(err))
		}
	}
}

type bufferHolder struct {
	buffer []byte
}

func newBufferHolder(size int) *bufferHolder {
	return &bufferHolder{
		buffer: make([]byte, size),
	}
}
