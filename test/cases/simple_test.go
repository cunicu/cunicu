package cases_test

import (
	"riasc.eu/wice/test/net"
	"riasc.eu/wice/test/nodes"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Context("simple", func() {

	Specify("any", func() {
		Run(net.Simple,
			&nodes.AgentParams{},
			&net.NetworkParams{
				NumAgents: 2,
			},
		)
	})

	Specify("any IPv6", func() {
		Run(net.Simple,
			&nodes.AgentParams{
				Arguments: []interface{}{
					"--ice-network-type", "udp6",
				},
			},
			&net.NetworkParams{
				NumAgents: 2,
			},
		)
	})

	Specify("host IPv4", func() {
		Run(net.Simple,
			&nodes.AgentParams{
				Arguments: []interface{}{
					"--ice-network-type", "udp4",
					"--ice-candidate-type", "srflx",
				},
			},
			&net.NetworkParams{
				NumAgents: 2,
			},
		)
	})

	Specify("host IPv6", func() {
		Run(net.Simple,
			&nodes.AgentParams{
				Arguments: []interface{}{
					"--ice-network-type", "udp6",
					"--ice-candidate-type", "srflx",
				},
			},
			&net.NetworkParams{
				NumAgents: 2,
			},
		)
	})

	Specify("srflx", func() {
		Run(net.Simple,
			&nodes.AgentParams{
				Arguments: []interface{}{
					"--ice-network-type", "udp4",
					"--ice-candidate-type", "srflx",
				},
			},
			&net.NetworkParams{
				NumAgents: 2,
			},
		)
	})

	Specify("relay", func() {
		Run(net.Simple,
			&nodes.AgentParams{
				Arguments: []interface{}{
					"--ice-network-type", "udp4",
					"--ice-candidate-type", "relay",
				},
			},
			&net.NetworkParams{
				NumAgents: 2,
			},
		)
	})
})
