package main

import (
	"log"
	"net"
	"os"

	"riasc.eu/wice/tc_test/bpf"

	"github.com/cilium/ebpf/rlimit"
	tc "github.com/florianl/go-tc"
)

func main() {
	// Allow the current process to lock memory for eBPF resources.
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatal(err)
	}

	if len(os.Args) < 2 {
		log.Fatalf("usage: %s INTF [INTF...]", os.Args[0])
	}

	intfNames := os.Args[1:]

	tcnl, err := tc.Open(&tc.Config{})
	if err != nil {
		log.Fatalf("Failed to open rtnetlink socket: %v\n", err)
	}
	defer tcnl.Close()

	objs, err := bpf.Load()
	if err != nil {
		log.Fatalf("Failed to load BPF code: %v\n", err)
	}

	addr := &net.UDPAddr{
		IP:   net.ParseIP("10.211.55.2"),
		Port: 1234,
	}

	addr2 := &net.UDPAddr{
		IP:   net.ParseIP("10.211.55.2"),
		Port: 1235,
	}

	if err := objs.Maps.EgressMap.AddEntry(addr, &bpf.MapStateEntry{
		ChannelId: 0xAABB,
		Lport:     2222,
	}); err != nil {
		log.Fatalf("Failed to add entry: %s", err)
	}

	if err := objs.Maps.EgressMap.AddEntry(addr2, &bpf.MapStateEntry{
		ChannelId: 0,
		Lport:     3333,
	}); err != nil {
		log.Fatalf("Failed to add entry: %s", err)
	}

	if err := objs.Maps.SettingsMap.EnableDebug(); err != nil {
		log.Fatalf("Failed to enable debugging: %s", err)
	}

	for _, intfName := range intfNames {
		intf, err := net.InterfaceByName(intfName)
		if err != nil {
			log.Fatalf("could not get interface ID: %v\n", err)
		}

		if err := bpf.AttachTCFilters(tcnl, intf.Index, objs); err != nil {
			log.Fatalf("failed to attach BPF filter: %s", err)
		}
	}
}
