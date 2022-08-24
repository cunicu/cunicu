package options

import (
	"riasc.eu/wice/test/nodes"
)

type ExtraArgs []any

func (ea ExtraArgs) Apply(a *nodes.Agent) {
	a.ExtraArgs = append(a.ExtraArgs, ea...)
}
