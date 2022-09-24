package config

import (
	"github.com/knadh/koanf"
	"github.com/stv0g/cunicu/pkg/util"
	"golang.org/x/exp/maps"
)

type Change struct {
	Old any
	New any
}

func DiffConfig(old, new *koanf.Koanf) map[string]Change {
	oldMap := old.Raw()
	newMap := new.Raw()

	return diff(oldMap, newMap)
}

func diff(old, new map[string]any) map[string]Change {
	added, removed, kept := util.DiffSlice(
		maps.Keys(old),
		maps.Keys(new),
	)

	changes := map[string]Change{}

	for _, key := range added {
		newValue := new[key]

		changes[key] = Change{
			New: newValue,
		}
	}

	for _, key := range removed {
		oldValue := old[key]

		changes[key] = Change{
			Old: oldValue,
		}
	}

	for _, key := range kept {
		oldStruct, oldIsStruct := old[key].(map[string]any)
		newStruct, newIsStruct := new[key].(map[string]any)

		if oldIsStruct && newIsStruct {
			for skey, chg := range diff(oldStruct, newStruct) {
				changes[key+"."+skey] = chg
			}
		} else {
			if old[key] != new[key] {
				changes[key] = Change{
					Old: old[key],
					New: new[key],
				}
			}
		}
	}

	return changes
}
