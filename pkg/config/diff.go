package config

import (
	"reflect"

	"github.com/stv0g/cunicu/pkg/util"
	"golang.org/x/exp/maps"
)

type Change struct {
	Old any
	New any
}

func DiffSettings(oldSettings, newSettings *Settings) map[string]Change {
	oldMap := Map(oldSettings, "koanf")
	newMap := Map(newSettings, "koanf")

	return diff(oldMap, newMap)
}

func diff(oldSettings, newSettings map[string]any) map[string]Change {
	added, removed, kept := util.SliceDiff(
		maps.Keys(oldSettings),
		maps.Keys(newSettings),
	)

	changes := map[string]Change{}

	for _, key := range added {
		newValue := newSettings[key]

		changes[key] = Change{
			New: newValue,
		}
	}

	for _, key := range removed {
		oldValue := oldSettings[key]

		changes[key] = Change{
			Old: oldValue,
		}
	}

	for _, key := range kept {
		oldStruct, oldIsStruct := oldSettings[key].(map[string]any)
		newStruct, newIsStruct := newSettings[key].(map[string]any)

		if oldIsStruct && newIsStruct {
			for skey, chg := range diff(oldStruct, newStruct) {
				changes[key+"."+skey] = chg
			}
		} else if !reflect.DeepEqual(oldSettings[key], newSettings[key]) {
			changes[key] = Change{
				Old: oldSettings[key],
				New: newSettings[key],
			}
		}
	}

	return changes
}
