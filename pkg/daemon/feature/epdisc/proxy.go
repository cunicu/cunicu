// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

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
