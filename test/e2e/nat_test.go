package e2e_test

import (
	"testing"

	"riasc.eu/wice/test/e2e/net"
)

func TestNAT(t *testing.T) {
	RunTest(t,
		net.NAT,
		&net.NetworkParams{
			NumAgents: 2,
		},
		[]interface{}{},
	)
}
