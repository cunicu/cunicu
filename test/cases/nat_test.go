package cases_test

import (
	"riasc.eu/wice/test/net"
	"riasc.eu/wice/test/nodes"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Context("nat", func() {
	Specify("nat", func() {
		Skip("p2p not yet supported")

		Run(net.NAT,
			&nodes.AgentParams{},
			&net.NetworkParams{
				NumAgents: 2,
			},
		)
	})
})
