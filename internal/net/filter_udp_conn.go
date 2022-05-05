package net

import (
	"encoding/hex"
	"fmt"
	"net"
	"time"

	"github.com/cilium/ebpf"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/mdlayher/socket"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
	"golang.org/x/net/bpf"
	"golang.org/x/sys/unix"
)

const (
	// Socket option to attach a classic BPF program to the socket for
	// use as a filter of incoming packets.
	SO_ATTACH_FILTER int = 26

	// Socket option to attach an extended BPF program to the socket for
	// use as a filter of incoming packets.
	SO_ATTACH_BPF int = 50
)

// Filter represents a classic BPF filter program that can be applied to a socket
type Filter []bpf.Instruction

type FilteredUDPConn struct {
	conn4 *socket.Conn
	conn6 *socket.Conn

	running4 bool
	running6 bool

	packets chan packet

	localPort int

	logger *zap.Logger
}

func (f *FilteredUDPConn) LocalAddr() net.Addr {
	return &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: f.localPort,
	}
}

type packet struct {
	N       int
	Address unix.Sockaddr
	Buffer  []byte
	Error   error
}

func (f *FilteredUDPConn) read(conn *socket.Conn, running *bool) {
	*running = true

	for {
		buf := make([]byte, 1500)
		if n, ra, err := conn.Recvfrom(buf, 0); err == nil {
			buf = buf[:n]
			f.packets <- packet{n, ra, buf, err}
		} else {
			f.packets <- packet{n, ra, buf, err}
			break
		}
	}

	*running = false
}

func (f *FilteredUDPConn) ReadFrom(buf []byte) (n int, addr net.Addr, err error) {
	// Wait for the next packet either from the IPv4/IPv6 connection
	pkt := <-f.packets
	if err, ok := pkt.Error.(net.Error); ok && err.Timeout() {
		return -1, nil, err
	}

	var ip net.IP
	var decoder gopacket.Decoder

	if sa, isIPv6 := pkt.Address.(*unix.SockaddrInet6); isIPv6 {
		ip = sa.Addr[:]
		decoder = layers.LayerTypeUDP
	} else if sa, isIPv4 := pkt.Address.(*unix.SockaddrInet4); isIPv4 {
		decoder = layers.LayerTypeIPv4
		ip = sa.Addr[:]
	} else {
		return -1, nil, fmt.Errorf("received invalid address family")
	}

	f.logger.Debug("Received packet",
		zap.Any("remote_address", ip),
		zap.Any("buf", hex.EncodeToString(pkt.Buffer)),
		zap.Any("decoder", decoder))

	packet := gopacket.NewPacket(pkt.Buffer, decoder, gopacket.DecodeOptions{
		Lazy:   true,
		NoCopy: true,
	})

	logWr := zapio.Writer{
		Log:   f.logger,
		Level: zap.DebugLevel,
	}
	logWr.Write([]byte(packet.Dump()))

	transport := packet.TransportLayer()
	if transport == nil {
		return -1, nil, fmt.Errorf("failed to decode packet")
	}

	udp, ok := transport.(*layers.UDP)
	if !ok {
		return -1, nil, fmt.Errorf("invalid layer type")
	}

	pl := packet.ApplicationLayer()
	n = len(pl.Payload())

	copy(buf[:n], pl.Payload()[:])

	rUDPAddr := &net.UDPAddr{
		IP:   ip,
		Port: int(udp.SrcPort),
	}

	return n, rUDPAddr, nil
}

func (f *FilteredUDPConn) WriteTo(buf []byte, rAddr net.Addr) (n int, err error) {
	f.logger.Info("helhelpdjklfhsdfkjhsldkfjghsdlkfgh")

	rUDPAddr, ok := rAddr.(*net.UDPAddr)
	if !ok {
		return -1, fmt.Errorf("invalid address type")
	}

	buffer := gopacket.NewSerializeBuffer()
	payload := gopacket.Payload(buf)

	udp := &layers.UDP{
		SrcPort: layers.UDPPort(f.localPort),
		DstPort: layers.UDPPort(rUDPAddr.Port),
	}

	var rSockAddr unix.Sockaddr
	var nwLayer gopacket.NetworkLayer
	var conn *socket.Conn

	isIPv6 := rUDPAddr.IP.To4() == nil

	f.logger.Info("Send packet", zap.String("addr", rUDPAddr.String()))

	if isIPv6 {
		sa := &unix.SockaddrInet6{}
		copy(sa.Addr[:], rUDPAddr.IP.To16())

		conn = f.conn6
		rSockAddr = sa
		nwLayer = &layers.IPv6{
			SrcIP: net.IPv6zero,
			DstIP: rUDPAddr.IP,
		}
	} else {
		sa := &unix.SockaddrInet4{}
		copy(sa.Addr[:], rUDPAddr.IP.To4())

		conn = f.conn4
		rSockAddr = sa
		nwLayer = &layers.IPv4{
			SrcIP: net.IPv4zero,
			DstIP: rUDPAddr.IP,
		}
	}

	if err := udp.SetNetworkLayerForChecksum(nwLayer); err != nil {
		return -1, fmt.Errorf("failed to set network layer for checksum: %w", err)
	}

	seropts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}
	if err := gopacket.SerializeLayers(buffer, seropts, udp, payload); err != nil {
		return -1, fmt.Errorf("failed serialize packet: %s", err)
	}

	bufser := buffer.Bytes()

	f.logger.Debug("Sending packet",
		zap.Any("remote_address", rSockAddr),
		zap.Any("buf", hex.EncodeToString(buf)))

	return 0, conn.Sendto(bufser, rSockAddr, 0)
}

func (f *FilteredUDPConn) ApplyFilter(prog *ebpf.Program) error {
	// Attach filter program
	if err := f.conn4.SetsockoptInt(unix.SOL_SOCKET, SO_ATTACH_BPF, prog.FD()); err != nil {
		return fmt.Errorf("failed setsockopt(fd, SOL_SOCKET, SO_ATTACH_BPF): %w", err)
	}

	if err := f.conn6.SetsockoptInt(unix.SOL_SOCKET, SO_ATTACH_BPF, prog.FD()); err != nil {
		return fmt.Errorf("failed setsockopt(fd, SOL_SOCKET, SO_ATTACH_BPF): %w", err)
	}

	return nil
}

func (f *FilteredUDPConn) setDeadlines(t time.Time, g func(*socket.Conn, time.Time) error) error {
	if err := g(f.conn4, t); err != nil {
		return fmt.Errorf("v4: %w", err)
	}

	if err := g(f.conn6, t); err != nil {
		return fmt.Errorf("v6: %w", err)
	}

	if !f.running4 {
		go f.read(f.conn4, &f.running4)
	}

	if !f.running6 {
		go f.read(f.conn6, &f.running6)
	}

	return nil
}

func (f *FilteredUDPConn) SetDeadline(t time.Time) error {
	return f.setDeadlines(t, (*socket.Conn).SetDeadline)
}

func (f *FilteredUDPConn) SetReadDeadline(t time.Time) error {
	return f.setDeadlines(t, (*socket.Conn).SetReadDeadline)
}

func (f *FilteredUDPConn) SetWriteDeadline(t time.Time) error {
	return f.setDeadlines(t, (*socket.Conn).SetWriteDeadline)
}

func (f *FilteredUDPConn) Close() error {
	if err := f.conn4.Close(); err != nil {
		return fmt.Errorf("v4: %w", err)
	}

	if err := f.conn6.Close(); err != nil {
		return fmt.Errorf("v4: %w", err)
	}

	return nil
}

func NewFilteredUDPConn(lPort int) (*FilteredUDPConn, error) {
	var err error

	f := &FilteredUDPConn{
		localPort: lPort,
		logger:    zap.L().Named("fuc"),
	}

	// SOCK_RAW sockets on Linux can only listen on a single address family (IPv4/IPv6)
	// This is different from normal SOCK_STREAM/SOCK_DGRAM sockets which for the case
	// of AF_INET6 also automatically listen on AF_INET.
	// Hence we need to open two independent sockets here.

	if f.conn4, err = socket.Socket(unix.AF_INET, unix.SOCK_RAW, unix.IPPROTO_UDP, "raw_udp4", nil); err != nil {
		return nil, fmt.Errorf("failed to open v4 raw socket: %w", err)
	}

	if f.conn6, err = socket.Socket(unix.AF_INET6, unix.SOCK_RAW, unix.IPPROTO_UDP, "raw_udp6", nil); err != nil {
		return nil, fmt.Errorf("failed to open v6 raw socket: %w", err)
	}

	// fuc.conn4.Bind(&unix.SockaddrInet4{
	// 	Port: fuc.localPort,
	// 	Addr: ,
	// })

	f.packets = make(chan packet)

	go f.read(f.conn4, &f.running4)
	go f.read(f.conn6, &f.running6)

	return f, nil
}
