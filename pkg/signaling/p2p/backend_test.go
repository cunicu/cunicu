package p2p_test

import (
	"testing"

	"riasc.eu/wice/internal/test"
)

func TestMain(m *testing.M) {
	test.Main(m)
}

func TestBackendP2P(t *testing.T) {
	test.TestBackend(t, "p2p:?private=true&mdns=true", 2)
}
