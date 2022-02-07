package e2e_test

import (
	"testing"

	g "github.com/stv0g/gont/pkg"
	gopt "github.com/stv0g/gont/pkg/options"
	"riasc.eu/wice/test/e2e/net"
)

func TestSimple(t *testing.T) {
	RunTest(t,
		net.Simple,
		&net.NetworkParams{
			NumAgents:      2,
			NetworkOptions: []g.Option{gopt.Persistent(true)},
		},
		[]interface{}{
			"--backend", "p2p:?private=true",
			// Limititing ourself to IPv4 network types
			// "--ice-network-type", "tcp4",
			// "--ice-network-type", "udp4",

			// "--ice-candidate-type", "relay",
			// "--proxy", "user",
		},
	)
}

func TestSimpleUser(t *testing.T) {
	RunTest(t,
		net.Simple,
		&net.NetworkParams{
			NumAgents: 2,
		},
		[]interface{}{
			"--proxy", "user",
		},
	)
}

func TestSimpleIPv6(t *testing.T) {
	RunTest(t,
		net.Simple,
		&net.NetworkParams{
			NumAgents: 3,
		},
		[]interface{}{
			"--proxy", "user",
			"--ice-network-type", "tcp6",
			"--ice-network-type", "udp6",
		},
	)
}

func TestSimpleTURN(t *testing.T) {
	RunTest(t,
		net.Simple,
		&net.NetworkParams{
			NumAgents: 3,
		},
		[]interface{}{
			"--proxy", "user",
			"--ice-candidate-type", "relay",
		},
	)
}
