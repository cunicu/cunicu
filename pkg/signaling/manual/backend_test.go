package manual_test

import (
	"testing"

	"riasc.eu/wice/internal/test"

	_ "riasc.eu/wice/pkg/signaling/manual"
)

func TestBackendInProcess(t *testing.T) {
	test.TestBackend(t, "manual", 10)
}
