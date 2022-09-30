// Package feat contains several sub-packages each implementing a dedicated feature.
package daemon

import (
	"golang.org/x/exp/slices"
)

var (
	Features = map[string]*FeaturePlugin{}
	plugins  []*FeaturePlugin
)

type FeaturePlugin struct {
	Name        string
	Description string

	New   func(i *Interface) (Feature, error)
	Order int
}

type SyncableFeature interface {
	Sync() error
}

type Feature interface {
	Start() error
	Close() error
}

func SortedFeatures() []*FeaturePlugin {
	if plugins == nil {
		for name, feat := range Features {
			feat.Name = name
			plugins = append(plugins, feat)
		}
	}

	slices.SortFunc(plugins, func(a, b *FeaturePlugin) bool {
		return a.Order < b.Order
	})

	return plugins
}
