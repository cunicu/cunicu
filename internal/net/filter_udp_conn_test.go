package net_test

import (
	"bytes"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/asm"
	"github.com/pion/stun"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"riasc.eu/wice/internal"
	netx "riasc.eu/wice/internal/net"
	"riasc.eu/wice/pkg/proxy"
)

func bpfSTUNTrafficOnPort(lPort int) asm.Instructions {
	return asm.Instructions{
		// LoadAbs() requires ctx in R6
		asm.Mov.Reg(asm.R6, asm.R1),

		// Offset of transport header from start of packet
		// IPv6 raw sockets do not include the network layer
		// so this is 0 by default
		asm.LoadImm(asm.R7, 0, asm.DWord),

		// r1 has ctx
		// r0 = ctx[16] (aka protocol)
		asm.LoadMem(asm.R0, asm.R1, 16, asm.Word),

		// Perhaps IPv6? Then skip the IPv4 part..
		asm.LoadImm(asm.R2, int64(unix.ETH_P_IPV6), asm.DWord),
		asm.HostTo(asm.BE, asm.R2, asm.Half),
		asm.JEq.Reg(asm.R0, asm.R2, "load"),

		// Transport layer starts after 20 Byte IPv4 header
		// TODO: use IHL field to account for IPv4 options
		asm.LoadImm(asm.R7, 20, asm.DWord),

		// Load UDP destination port
		asm.LoadInd(asm.R0, asm.R7, 2, asm.Half).Sym("load"),

		// Skip if is not matching our listen port
		asm.JNE.Imm(asm.R0, int32(lPort), "skip"),

		// Load STUN Magic Cookie from UDP payload
		asm.LoadInd(asm.R0, asm.R7, 12, asm.Word),

		// Skip if it is not the well know value
		asm.JNE.Imm(asm.R0, int32(proxy.StunMagicCookie), "skip"),

		asm.Mov.Imm(asm.R0, -1).Sym("exit"),
		asm.Return(),

		asm.Mov.Imm(asm.R0, 0).Sym("skip"),
		asm.Return(),
	}
}

func exchangePacket(t *testing.T, expectValid bool, makeInvalid bool, netw string, port int, receiverConn net.PacketConn) {
	var msg = stun.MustBuild(stun.TransactionID, stun.BindingRequest)

	if makeInvalid {
		msg.Raw[4] = 0 // we destroy STUNs magic cookie here
	}

	var addr string
	if netw == "udp6" {
		addr = "[::1]"
	} else {
		addr = "127.0.0.1"
	}

	senderConn, err := net.Dial(netw, fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		t.Fatalf("failed to dial IPv4")
	}
	defer senderConn.Close()

	if _, err := senderConn.Write(msg.Raw); err != nil {
		t.Fatalf("failed to send packet: %s", err)
	}

	if !expectValid {
		// Invalid messages should never pass the filter
		// So we set a timeout here and assert that the timeout will expire
		receiverConn.SetReadDeadline(time.Now().Add(5 * time.Millisecond))
	}

	recvMsg := make([]byte, 1024)
	n, _, err := receiverConn.ReadFrom(recvMsg)
	if expectValid {
		if err != nil {
			t.Fatalf("failed to read from connection: %s", err)
		}

		if n != len(msg.Raw) {
			t.Fatal("mismatching length")
		}

		if bytes.Compare(msg.Raw, recvMsg[:n]) != 0 {
			t.Fatal("mismatching contents")
		}
	} else {
		if err, ok := err.(net.Error); !ok || !err.Timeout() {
			t.Fatalf("received some data when we should not: %v (%d)", err, n)
		}
	}
}

func TestFilterUDPConn(t *testing.T) {
	internal.SetupLogging(zap.DebugLevel, "")

	la := net.UDPAddr{
		IP:   nil,
		Port: 12345,
	}

	// We are opening a standard UDP socket here which does not get used
	// Its only there to avoid the system to send ICMP port unreachable messages
	// as well as to test of the filtered UDP socket can run alongside already listening sockets
	// without raising EADDRINUSE.
	_, err := net.ListenUDP("udp", &la)
	if err != nil {
		t.Fatalf("failed to open UDP socket: %s", err)
	}

	spec := ebpf.ProgramSpec{
		Type:         ebpf.SocketFilter,
		License:      "GPL",
		Instructions: bpfSTUNTrafficOnPort(la.Port),
	}

	if testing.Verbose() {
		t.Logf("Instructions:\n%v\n", spec.Instructions)
	}

	prog, err := ebpf.NewProgramWithOptions(&spec, ebpf.ProgramOptions{
		LogLevel: 6, // TODO take configured log-level from args
	})
	if err != nil {
		t.Fatalf("Failed to create eBPF program: %s", err)
	}

	f, err := netx.NewFilteredUDPConn(la.Port)
	if err != nil {
		t.Fatalf("Failed to create filtered UDP connection: %s", err)
	}

	if err := f.ApplyFilter(prog); err != nil {
		t.Fatalf("failed to apply eBPF filter: %s", err)
	}

	// Send valid packets
	for i := 0; i < 2; i++ {
		exchangePacket(t, true, false, "udp4", 12345, f)
		exchangePacket(t, true, false, "udp6", 12345, f)
	}

	// Send non-STUN packets to filtered connection
	for i := 0; i < 2; i++ {
		exchangePacket(t, false, true, "udp4", 12345, f)
		exchangePacket(t, false, true, "udp6", 12345, f)
	}

	// Send STUN packets to another port
	for i := 0; i < 2; i++ {
		exchangePacket(t, false, false, "udp4", 11111, f)
		exchangePacket(t, false, false, "udp6", 11111, f)
	}

	// Send valid packets (again)
	f.SetReadDeadline(time.Time{})
	for i := 0; i < 2; i++ {
		exchangePacket(t, true, false, "udp4", 12345, f)
		exchangePacket(t, true, false, "udp6", 12345, f)
	}
}
