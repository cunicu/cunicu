package k8s_test

import (
	"os"
	"testing"

	"riasc.eu/wice/internal/test"
)

func TestBackend(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skipf("Kubernetes tests are not yet supported in CI")
	}

	test.TestBackend(t, "k8s:?node=red")
}
