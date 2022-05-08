package net_test

import (
	"fmt"
	"net"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/asm"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pion/stun"
	"golang.org/x/sys/unix"
	"kernel.org/pub/linux/libs/security/libcap/cap"
	netx "riasc.eu/wice/internal/net"
	"riasc.eu/wice/internal/util"
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

var _ = Describe("FilteredUDPConn", Ordered, func() {
	var c net.Conn
	var f *netx.FilteredUDPConn

	la := net.UDPAddr{
		IP:   nil,
		Port: 12345,
	}

	BeforeAll(func() {
		if !util.HasCapabilities(cap.NET_RAW) {
			Skip("Insufficient privileges")
		}

		// We are opening a standard UDP socket here which does not get used
		// Its only there to avoid the system to send ICMP port unreachable messages
		// as well as to test of the filtered UDP socket can run alongside already listening sockets
		// without raising EADDRINUSE.
		var err error
		c, err = net.ListenUDP("udp", &la)
		Expect(err).To(Succeed(), "Failed to open socket: %s", err)

		spec := ebpf.ProgramSpec{
			Type:         ebpf.SocketFilter,
			License:      "GPL",
			Instructions: bpfSTUNTrafficOnPort(la.Port),
		}

		prog, err := ebpf.NewProgramWithOptions(&spec, ebpf.ProgramOptions{LogLevel: 6})
		Expect(err).To(Succeed(), "Failed to create eBPF program: %s", err)

		f, err = netx.NewFilteredUDPConn(la.Port)
		Expect(err).To(Succeed(), "Failed to create filtered UDP connection: %s", err)
		Expect(f.ApplyFilter(prog)).To(Succeed(), "failed to apply eBPF filter: %s", err)
	})

	DescribeTable("Send valid packets", FlakeAttempts(2), func(shouldSucceed, makeInvalid bool, netw string, port int) {
		msg := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
		if makeInvalid {
			msg.Raw[4] = 0 // we destroy STUNs magic cookie here
		}

		var addr string
		if netw == "udp6" {
			addr = "[::1]"
		} else {
			addr = "127.0.0.1"
		}

		s, err := net.Dial(netw, fmt.Sprintf("%s:%d", addr, port))
		Expect(err).To(Succeed(), "failed to dial IPv4: %s", err)
		defer s.Close()

		_, err = s.Write(msg.Raw)
		Expect(err).To(Succeed(), "failed to send packet: %s", err)

		// Invalid messages should never pass the filter
		// So we set a timeout here and assert that the timeout will expire
		if shouldSucceed {
			err = f.SetDeadline(time.Time{}) // Reset
		} else {
			err = f.SetReadDeadline(time.Now().Add(5 * time.Millisecond))
		}
		Expect(err).To(Succeed())

		recvMsg := make([]byte, 1024)
		n, _, err := f.ReadFrom(recvMsg)
		if shouldSucceed {
			Expect(err).To(Succeed(), "failed to read from connection: %s", err)
			Expect(n).To(Equal(len(msg.Raw)), "mismatching length")
			Expect(msg.Raw).To(Equal(recvMsg[:n]), "mismatching contents")
		} else {
			err, isNetError := err.(net.Error)
			Expect(isNetError).To(BeTrue(), "invalid error type: %s", err)
			Expect(err.Timeout()).To(BeTrue(), "error is not a timeout")
		}
	},
		Entry("Valid IPv4", true, false, "udp4", 12345),
		Entry("Valid IPv6", true, false, "udp6", 12345),
		Entry("Non-STUN IPv4", false, true, "udp4", 12345),
		Entry("Non-STUN IPv6", false, true, "udp6", 12345),
		Entry("STUN to different port", false, false, "udp4", 11111),
		Entry("STUN to different port", false, false, "udp6", 11111),
		Entry("Valid IPv4 (again)", true, false, "udp4", 12345),
		Entry("Valid IPv6 (again)", true, false, "udp6", 12345),
	)

	AfterAll(func() {
		Expect(c.Close()).To(Succeed())
		Expect(f.Close()).To(Succeed())
	})
})
