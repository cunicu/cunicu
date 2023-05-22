package epdisc

import (
	"errors"
	"io"
)

var errMismatchingEndpoints = errors.New("mismatching EPs")

type Proxy interface {
	io.Closer
}

type ProxyConn struct {
	Proxy

	io.Closer
}
