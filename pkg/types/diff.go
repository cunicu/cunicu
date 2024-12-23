// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"reflect"

	kmaps "github.com/knadh/koanf/maps"

	mapsx "cunicu.li/cunicu/pkg/types/maps"
	slicesx "cunicu.li/cunicu/pkg/types/slices"
)

type Change struct {
	Old any
	New any
}

func DiffMap(oldMap, newMap map[string]any) map[string]Change {
	oldMap = kmaps.Unflatten(oldMap, ".")
	newMap = kmaps.Unflatten(newMap, ".")

	return diffMap(oldMap, newMap)
}

func diffMap(oldMap, newMap map[string]any) map[string]Change {
	added, removed, kept := slicesx.Diff(
		mapsx.Keys(oldMap),
		mapsx.Keys(newMap),
	)

	changes := map[string]Change{}

	for _, key := range added {
		newValue := newMap[key]

		changes[key] = Change{
			New: newValue,
		}
	}

	for _, key := range removed {
		oldValue := oldMap[key]

		changes[key] = Change{
			Old: oldValue,
		}
	}

	for _, key := range kept {
		oldSub, oldIsMap := oldMap[key].(map[string]any)
		newSub, newIsMap := newMap[key].(map[string]any)

		// Descent if both keys are maps
		if oldIsMap && newIsMap {
			for sKey, chg := range diffMap(oldSub, newSub) {
				changes[key+"."+sKey] = chg
			}
		} else if !reflect.DeepEqual(oldMap[key], newMap[key]) {
			changes[key] = Change{
				Old: oldMap[key],
				New: newMap[key],
			}
		}
	}

	return changes
}
