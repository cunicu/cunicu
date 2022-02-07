package e2e_test

import (
	"testing"

	"riasc.eu/wice/test/e2e/net"
)

func TestFiltered(t *testing.T) {
	RunTest(t,
		net.Filtered,
		&net.NetworkParams{
			NumAgents: 2,
		},
		[]interface{}{},
	)
}
