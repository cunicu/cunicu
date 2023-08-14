// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package nodes

import (
	"net/url"

	g "cunicu.li/gont/v2/pkg"
)

type RelayNode interface {
	Node

	WaitReady() error
	URLs() []url.URL

	// Options
	Apply(i *g.Interface)
}
