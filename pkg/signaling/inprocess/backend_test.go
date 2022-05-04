package inprocess_test

import (
	"testing"

	"riasc.eu/wice/internal/test"

	_ "riasc.eu/wice/pkg/signaling/inprocess"
)

func TestBackendInProcess(t *testing.T) {
	test.TestBackend(t, "inprocess", 10)
}
