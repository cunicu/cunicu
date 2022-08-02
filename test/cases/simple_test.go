//go:build linux

package cases_test

import (
	"riasc.eu/wice/test/net"
	"riasc.eu/wice/test/nodes"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Context("simple", func() {

	Context("any", func() {
		Specify("ipv4", func() {
			Run(net.Simple,
				&nodes.AgentParams{},
				&net.NetworkParams{
					NumAgents: 2,
				},
			)
		})

		Specify("ipv6", func() {
			Run(net.Simple,
				&nodes.AgentParams{
					Arguments: []any{
						"--ice-network-type", "udp6",
					},
				},
				&net.NetworkParams{
					NumAgents: 2,
				},
			)
		})
	})

	Context("host", func() {
		Specify("ipv4", func() {
			Run(net.Simple,
				&nodes.AgentParams{
					Arguments: []any{
						"--ice-network-type", "udp4",
						"--ice-candidate-type", "srflx",
					},
				},
				&net.NetworkParams{
					NumAgents: 2,
				},
			)
		})

		Specify("ipv6", func() {
			Run(net.Simple,
				&nodes.AgentParams{
					Arguments: []any{
						"--ice-network-type", "udp6",
						"--ice-candidate-type", "srflx",
					},
				},
				&net.NetworkParams{
					NumAgents: 2,
				},
			)
		})
	})

	Context("srflx", func() {
		Specify("ipv4", func() {
			Run(net.Simple,
				&nodes.AgentParams{
					Arguments: []any{
						"--ice-network-type", "udp4",
						"--ice-candidate-type", "srflx",
					},
				},
				&net.NetworkParams{
					NumAgents: 2,
				},
			)
		})

		Specify("ipv6", func() {
			Run(net.Simple,
				&nodes.AgentParams{
					Arguments: []any{
						"--ice-network-type", "udp6",
						"--ice-candidate-type", "srflx",
					},
				},
				&net.NetworkParams{
					NumAgents: 2,
				},
			)
		})
	})

	Context("relay", func() {
		Specify("ipv4", func() {
			Run(net.Simple,
				&nodes.AgentParams{
					Arguments: []any{
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
})
