package config_test

import (
	"testing"

	"riasc.eu/wice/internal/config"
)

func TestParseArgsUser(t *testing.T) {
	config, err := config.Parse("--wg-userspace")
	if err != nil {
		t.Errorf("err got %v, want nil", err)
	}

	if !config.GetBool("wg.userspace") {
		t.Fail()
	}
}

func TestParseArgsBackends(t *testing.T) {
	config, err := config.Parse("--backend", "k8s", "--backend", "p2p")
	if err != nil {
		t.Errorf("err got %v, want nil", err)
	}

	if len(config.Backends) != 2 {
		t.FailNow()
	}

	t.Logf("Backends: %+#v", config.Backends)

	if config.Backends[0].Scheme != "k8s" {
		t.Fail()
	}

	if config.Backends[1].Scheme != "p2p" {
		t.Fail()
	}
}
