// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

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
