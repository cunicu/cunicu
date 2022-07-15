package cases_test

import (
	"riasc.eu/wice/test/net"
	"riasc.eu/wice/test/nodes"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Context("filtered", func() {
	Specify("filtered", Pending, func() {
		Run(net.Filtered,
			&nodes.AgentParams{},
			&net.NetworkParams{
				NumAgents: 2,
			},
		)
	})
})
