package bpf

import (
	"fmt"

	"github.com/florianl/go-tc"
	"github.com/florianl/go-tc/core"
	"github.com/vishvananda/netlink/nl"
	"golang.org/x/sys/unix"
)

func AttachTCFilters(tcnl *tc.Tc, ifIndex int, objs *Objects) error {
	qdisc := tc.Object{
		Msg: tc.Msg{
			Family:  unix.AF_UNSPEC,
			Ifindex: uint32(ifIndex),
			Handle:  core.BuildHandle(tc.HandleRoot, 0x0000),
			Parent:  tc.HandleIngress,
			Info:    0,
		},
		Attribute: tc.Attribute{
			Kind: "clsact",
		},
	}

	if err := tcnl.Qdisc().Add(&qdisc); err != nil {
		return fmt.Errorf("could not assign clsact to %w", err)
	}

	m := map[int]uint32{
		objs.Programs.EgressFilter.FD():  core.BuildHandle(tc.HandleRoot, tc.HandleMinEgress),
		objs.Programs.IngressFilter.FD(): core.BuildHandle(tc.HandleRoot, tc.HandleMinIngress),
	}

	for fd, parent := range m {
		flags := uint32(nl.TCA_BPF_FLAG_ACT_DIRECT)
		fd2 := uint32(fd)

		filter := tc.Object{
			Msg: tc.Msg{
				Family:  unix.AF_UNSPEC,
				Ifindex: uint32(ifIndex),
				Handle:  0,
				Parent:  parent,
				Info:    0x300,
			},
			Attribute: tc.Attribute{
				Kind: "bpf",
				BPF: &tc.Bpf{
					FD:    &fd2,
					Flags: &flags,
				},
			},
		}
		if err := tcnl.Filter().Add(&filter); err != nil {
			return fmt.Errorf("failed to attach filter for eBPF program: %w", err)
		}
	}

	return nil
}
