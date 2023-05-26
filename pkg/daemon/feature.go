// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package daemon

import (
	"golang.org/x/exp/slices"
)

//nolint:gochecknoglobals
var (
	features = []*Feature{}
)

type Feature struct {
	New   func(i *Interface) (FeatureInterface, error)
	order int
}

type SyncableFeatureInterface interface {
	Sync() error
}

type FeatureInterface interface {
	Start() error
	Close() error
}

func RegisterFeature[I FeatureInterface](ctor func(i *Interface) (I, error), order int,
) func(*Interface) I {
	feature := &Feature{
		New: func(i *Interface) (FeatureInterface, error) {
			return ctor(i)
		},
		order: order,
	}

	features = append(features, feature)
	slices.SortFunc(features, func(a, b *Feature) bool { return a.order < b.order })

	return func(i *Interface) (q I) {
		if j, ok := i.features[feature]; ok {
			if p, ok := j.(I); ok {
				q = p
			}
		}
		return
	}
}
