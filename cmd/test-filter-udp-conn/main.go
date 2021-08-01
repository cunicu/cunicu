package main

import (
	"encoding/hex"
	"fmt"
	"net"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/asm"
	netx "riasc.eu/wice/internal/net"
)

const (
	StunMagicCookie uint32 = 0x2112A442
)

func main() {
	la := net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 12345,
	}

	spec := ebpf.ProgramSpec{
		Type:    ebpf.SocketFilter,
		License: "GPL",
		Instructions: asm.Instructions{
			asm.Mov.Reg(asm.R6, asm.R1), // LDABS requires ctx in R6
			asm.LoadAbs(-0x100000+22, asm.Half),
			asm.JNE.Imm(asm.R0, int32(la.Port), "skip"),
			asm.LoadAbs(-0x100000+32, asm.Word),
			asm.JNE.Imm(asm.R0, int32(StunMagicCookie), "skip"),
			asm.Mov.Imm(asm.R0, -1).Sym("exit"),
			asm.Return(),
			asm.Mov.Imm(asm.R0, 0).Sym("skip"),
			asm.Return(),
		},
	}

	fmt.Printf("Instructions:\n%v\n", spec.Instructions)

	prog, err := ebpf.NewProgramWithOptions(&spec, ebpf.ProgramOptions{
		LogLevel: 6, // TODO take configured log-level from args
	})
	if err != nil {
		panic(err)
	}

	fuc, err := netx.NewFilteredUDPConn(la)
	if err != nil {
		panic(err)
	}

	err = fuc.ApplyFilter(prog)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 1024)
	for {
		n, ra, err := fuc.ReadFrom(buf)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Bytes: %d\n", n)
		fmt.Printf("RA: %+v\n", ra)
		fmt.Printf("Bytes: %s\n", hex.EncodeToString(buf[:n]))
		fmt.Println()
	}
}
