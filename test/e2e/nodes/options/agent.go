// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package options

import (
	"github.com/stv0g/cunicu/test/e2e/nodes"
)

type ExtraArgs []any

func (ea ExtraArgs) Apply(a *nodes.Agent) {
	a.ExtraArgs = append(a.ExtraArgs, ea...)
}
