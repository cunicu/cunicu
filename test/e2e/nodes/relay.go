package nodes

import (
	"github.com/pion/ice/v2"
	g "github.com/stv0g/gont/v2/pkg"
)

type RelayNode interface {
	Node

	WaitReady() error
	URLs() []*ice.URL
	Username() string
	Password() string

	// Options
	Apply(i *g.Interface)
}
