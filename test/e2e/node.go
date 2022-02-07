//go:build linux

package e2e

import (
	g "github.com/stv0g/gont/pkg"
)

type Node interface {
	g.Node

	Start(args ...interface{}) error
	Stop() error
	Close() error
}
