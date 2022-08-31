package bpf

import (
	"fmt"
	"unsafe"

	"github.com/cilium/ebpf"
)

var stateMapInnerSpec = ebpf.MapSpec{
	Type:       ebpf.Hash,
	KeySize:    4,
	ValueSize:  uint32(unsafe.Sizeof(MapStateEntry{})),
	MaxEntries: 1 << 12,
}

func Load() (*Objects, error) {
	cs, err := loadBpf()
	if err != nil {
		return nil, fmt.Errorf("failed to load collection spec: %s", err)
	}

	for _, p := range cs.Programs {
		p.Type = ebpf.SchedCLS
	}

	for n, m := range cs.Maps {
		if n == "ingress_map" || n == "egress_map" {
			m.InnerMap = &stateMapInnerSpec
		}
	}

	objs := &bpfObjects{}
	if err := cs.LoadAndAssign(objs, nil); err != nil {
		return nil, fmt.Errorf("failed to load programs: %w", err)
	}

	return &Objects{
		Maps: Maps{
			SettingsMap: MapSettings{Map: objs.bpfMaps.SettingsMap},
			IngressMap:  MapState{Map: objs.bpfMaps.IngressMap},
			EgressMap:   MapState{Map: objs.bpfMaps.EgressMap},
		},
		Programs: objs.bpfPrograms,
	}, nil
}
