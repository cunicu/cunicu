package e2e

import (
	"github.com/pion/ice/v2"
	g "github.com/stv0g/gont/pkg"
)

type RelayNode interface {
	Node

	WaitReady() error
	IsReachable() bool
	URLs() []*ice.URL
	Username() string
	Password() string

	// Options
	Apply(i *g.Interface)
}
