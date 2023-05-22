package nodes

import (
	"github.com/pion/stun"
	g "github.com/stv0g/gont/v2/pkg"
)

type RelayNode interface {
	Node

	WaitReady() error
	URLs() []*stun.URI
	Username() string
	Password() string

	// Options
	Apply(i *g.Interface)
}
