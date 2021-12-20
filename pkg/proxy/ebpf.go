//go:build linux
// +build linux

package proxy

import (
	"fmt"
	"net"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/asm"
	"github.com/pion/ice/v2"

	icex "riasc.eu/wice/internal/ice"
	netx "riasc.eu/wice/internal/net"

	log "github.com/sirupsen/logrus"
)

type EBPFProxy struct {
	BaseProxy
}

func NewEBPFProxy(ident string, listenPort int, cb UpdateEndpointCb, conn net.Conn) (Proxy, error) {

	rUDPAddr := conn.RemoteAddr().(*net.UDPAddr)
	cb(rUDPAddr)

	return &EBPFProxy{
		BaseProxy: BaseProxy{
			Ident: ident,
		},
		// Conn: conn,
	}, nil
}

func (p *EBPFProxy) Type() ProxyType {
	return ProxyTypeEBPF
}

func SetupEBPFProxy(agentConfig *ice.AgentConfig, listenPort int) error {
	addr := net.UDPAddr{
		IP:   net.IPv4zero,
		Port: listenPort,
	}

	conn, err := netx.NewFilteredUDPConn(addr)
	if err != nil {
		return fmt.Errorf("failed to create filtered UDP connection: %w", err)
	}

	spec := ebpf.ProgramSpec{
		Type:    ebpf.SocketFilter,
		License: "Apache-2.0",
		Instructions: asm.Instructions{
			asm.Mov.Reg(asm.R6, asm.R1), // LDABS requires ctx in R6
			asm.LoadAbs(-0x100000+22, asm.Half),
			asm.JNE.Imm(asm.R0, int32(listenPort), "skip"),
			asm.LoadAbs(-0x100000+32, asm.Word),
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
		return fmt.Errorf("failed to create BPF program: %w", err)
	}

	if err = conn.ApplyFilter(prog); err != nil {
		return fmt.Errorf("failed to attach eBPF program to socket: %w", err)
	}

	agentConfig.UDPMux = icex.NewFilteredUDPMux(icex.FilteredUDPMuxParams{
		Logger: log.WithField("logger", "ice-mux"),
		Conn:   conn,
	})

	return nil
}

func (bpf *EBPFProxy) Close() error {
	return nil
}

func (bpf *EBPFProxy) UpdateEndpoint(addr *net.UDPAddr) error {
	return nil
}
