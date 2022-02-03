package ice

import (
	"fmt"
	"net"
	"time"

	"github.com/cilium/ebpf"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"go.uber.org/zap"
	"golang.org/x/net/bpf"

	"syscall"
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
	fd int

	localAddr net.UDPAddr

	logger *zap.Logger
}

func (fuc *FilteredUDPConn) LocalAddr() net.Addr {
	return &fuc.localAddr
}

func (fuc *FilteredUDPConn) ReadFrom(buf []byte) (n int, addr net.Addr, err error) {
	n, rAddr, err := syscall.Recvfrom(fuc.fd, buf, 0)
	if err != nil {
		return -1, nil, err
	}

	rAddrIn4, ok := rAddr.(*syscall.SockaddrInet4)
	if !ok {
		return -1, nil, fmt.Errorf("invalid address type")
	}

	packet := gopacket.NewPacket(buf[:n], layers.LayerTypeIPv4, gopacket.DecodeOptions{
		Lazy:   true,
		NoCopy: true,
	})

	transport := packet.TransportLayer()
	if transport == nil {
		return -1, nil, fmt.Errorf("failed to decode packet")
	}
	udp, ok := transport.(*layers.UDP)
	if !ok {
		return -1, nil, fmt.Errorf("invalid layer type")
	}
	pl := packet.ApplicationLayer()

	rUDPAddr := &net.UDPAddr{
		IP:   rAddrIn4.Addr[:],
		Port: int(udp.SrcPort),
	}

	n = len(pl.Payload())

	copy(buf[:n], pl.Payload()[:])

	// fuc.logger.Debug("Read data from socket",
	// 	zap.Any("ra", rUDPAddr),
	// 	zap.Int("len", n),
	// 	zap.String("buf", hex.EncodeToString(buf[:n])),
	// )

	return n, rUDPAddr, nil
}

func (fuc *FilteredUDPConn) WriteTo(buf []byte, rAddr net.Addr) (n int, err error) {
	rUDPAddr, ok := rAddr.(*net.UDPAddr)
	if !ok {
		return -1, fmt.Errorf("invalid address type")
	}

	rSockAddr := &syscall.SockaddrInet4{
		Port: 0,
	}
	copy(rSockAddr.Addr[:], rUDPAddr.IP.To4())

	buffer := gopacket.NewSerializeBuffer()
	payload := gopacket.Payload(buf)
	ip := &layers.IPv4{
		Version:  4,
		TTL:      64,
		SrcIP:    fuc.localAddr.IP,
		DstIP:    rUDPAddr.IP,
		Protocol: layers.IPProtocolUDP,
	}
	udp := &layers.UDP{
		SrcPort: layers.UDPPort(fuc.localAddr.Port),
		DstPort: layers.UDPPort(rUDPAddr.Port),
	}
	udp.SetNetworkLayerForChecksum(ip)
	seropts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}
	if err := gopacket.SerializeLayers(buffer, seropts, udp, payload); err != nil {
		return -1, fmt.Errorf("failed serialize packet: %s", err)
	}

	syscall.Sendto(fuc.fd, buffer.Bytes(), 0, rSockAddr)

	// fuc.logger.Debug("Written data to socket",
	// 	zap.Any("ra", rUDPAddr),
	// 	zap.Int("len", len(buf)),
	// 	zap.String("buf", hex.EncodeToString(buf)),
	// )

	return 0, nil
}

func (fuc *FilteredUDPConn) ApplyFilter(prog *ebpf.Program) error {
	// Attach filter program
	if err := syscall.SetsockoptInt(fuc.fd, syscall.SOL_SOCKET, SO_ATTACH_BPF, prog.FD()); err != nil {
		return fmt.Errorf("failed setsockopt(fd, SOL_SOCKET, SO_ATTACH_BPF): %w", err)
	}

	return nil
}

func (fuc *FilteredUDPConn) SetDeadline(t time.Time) error {
	// TODO
	return nil
}

func (fuc *FilteredUDPConn) SetReadDeadline(t time.Time) error {
	// TODO
	return nil
}

func (fuc *FilteredUDPConn) SetWriteDeadline(t time.Time) error {
	// TODO
	return nil
}

func (fuc *FilteredUDPConn) Close() error {
	// TODO
	return nil
}

func NewFilteredUDPConn(lAddr net.UDPAddr) (fuc *FilteredUDPConn, err error) {
	fuc = &FilteredUDPConn{
		localAddr: lAddr,
		logger:    zap.L().Named("fuc"),
	}

	// Open a raw socket
	fuc.fd, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_UDP)
	if err != nil {
		panic(err)
	}

	return fuc, nil
}
