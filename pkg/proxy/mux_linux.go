package proxy

import (
	"fmt"
	"net"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/asm"
	"golang.org/x/sys/unix"

	netx "riasc.eu/wice/internal/net"
)

func createFilteredSTUNConnection(listenPort int) (net.PacketConn, error) {
	conn, err := netx.NewFilteredUDPConn(listenPort)
	if err != nil {
		return nil, fmt.Errorf("failed to create filtered UDP connection: %w", err)
	}

	spec := ebpf.ProgramSpec{
		Type:    ebpf.SocketFilter,
		License: "Apache-2.0",
		Instructions: asm.Instructions{
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
			asm.JNE.Imm(asm.R0, int32(listenPort), "skip"),

			// Load STUN Magic Cookie from UDP payload
			asm.LoadInd(asm.R0, asm.R7, 12, asm.Word),

			// Skip if it is not the well know value
			asm.JNE.Imm(asm.R0, int32(StunMagicCookie), "skip"),

			asm.Mov.Imm(asm.R0, -1).Sym("exit"),
			asm.Return(),

			asm.Mov.Imm(asm.R0, 0).Sym("skip"),
			asm.Return(),
		},
	}
	prog, err := ebpf.NewProgramWithOptions(&spec, ebpf.ProgramOptions{
		LogLevel: 1, // TODO take configured log-level from args
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create BPF program: %w", err)
	}

	if err = conn.ApplyFilter(prog); err != nil {
		return nil, fmt.Errorf("failed to attach eBPF program to socket: %w", err)
	}

	return conn, nil
}
