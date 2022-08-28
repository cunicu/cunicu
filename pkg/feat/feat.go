// Package feat contains several sub-packages each implementing a dedicated feature.
package feat

type Feature interface {
	Start() error
	Close() error

	// Priority() int
	// Name() string
	// Description() string
}
