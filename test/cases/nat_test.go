package cases_test

import (
	"riasc.eu/wice/test/net"
	"riasc.eu/wice/test/nodes"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Context("nat", func() {
	Specify("nat", Pending, func() {
		Run(net.NAT,
			&nodes.AgentParams{},
			&net.NetworkParams{
				NumAgents: 2,
			},
		)
	})
})
