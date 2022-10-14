// Package feat contains several sub-packages each implementing a dedicated feature.
package daemon

import (
	"golang.org/x/exp/slices"
)

var (
	features       = map[string]*FeaturePlugin{}
	featuresSorted []*FeaturePlugin
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

func RegisterFeature(name, desc string, New func(i *Interface) (Feature, error), order int) {
	features[name] = &FeaturePlugin{
		Name:        name,
		Description: desc,
		New:         New,
		Order:       order,
	}
}

func SortedFeatures() []*FeaturePlugin {
	if featuresSorted == nil {
		for _, feat := range features {
			featuresSorted = append(featuresSorted, feat)
		}
	}

	slices.SortFunc(featuresSorted, func(a, b *FeaturePlugin) bool {
		return a.Order < b.Order
	})

	return featuresSorted
}
