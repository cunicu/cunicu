package config_test

import (
	"testing"

	"github.com/pion/ice/v2"
	"riasc.eu/wice/internal/config"
)

func TestParseArgsUser(t *testing.T) {
	config, err := config.Parse("--wg-user")
	if err != nil {
		t.Errorf("err got %v, want nil", err)
	}

	if !config.User {
		t.Fail()
	}
}

func TestParseArgsBackend(t *testing.T) {
	config, err := config.Parse("--backend", "k8s")
	if err != nil {
		t.Errorf("err got %v, want nil", err)
	}

	if config.Backends[0].Scheme != "k8s" {
		t.Fail()
	}
}

func TestParseArgsUrls(t *testing.T) {
	config, err := config.Parse("--url", "stun:stun.riasc.eu", "--url", "turn:turn.riasc.eu")
	if err != nil {
		t.Errorf("err got %v, want nil", err)
	}

	agentConfig, err := config.AgentConfig()
	if err != nil {
		t.FailNow()
	}

	if len(agentConfig.Urls) != 2 {
		t.Fail()
	}

	if agentConfig.Urls[0].Host != "stun.riasc.eu" {
		t.Fail()
	}

	if agentConfig.Urls[0].Scheme != ice.SchemeTypeSTUN {
		t.Fail()
	}

	if agentConfig.Urls[1].Host != "turn.riasc.eu" {
		t.Fail()
	}

	if agentConfig.Urls[1].Scheme != ice.SchemeTypeTURN {
		t.Fail()
	}
}

func TestParseArgsCandidateTypes(t *testing.T) {
	config, err := config.Parse("--ice-candidate-type", "host", "--ice-candidate-type", "relay")
	if err != nil {
		t.Errorf("err got %v, want nil", err)
	}

	agentConfig, err := config.AgentConfig()
	if err != nil {
		t.FailNow()
	}

	if len(agentConfig.CandidateTypes) != 2 {
		t.Fail()
	}

	if agentConfig.CandidateTypes[0] != ice.CandidateTypeHost {
		t.Fail()
	}

	if agentConfig.CandidateTypes[1] != ice.CandidateTypeRelay {
		t.Fail()
	}
}

func TestParseArgsInterfaceFilter(t *testing.T) {
	config, err := config.Parse("--interface-filter", "eth\\d+")
	if err != nil {
		t.Errorf("err got %v, want nil", err)
	}

	if !config.InterfaceFilter.Match([]byte("eth0")) {
		t.Fail()
	}

	if config.InterfaceFilter.Match([]byte("wifi0")) {
		t.Fail()
	}
}

func TestParseArgsInterfaceFilterFail(t *testing.T) {
	_, err := config.Parse("--interface-filter", "eth(")
	if err == nil {
		t.Fail()
	}
}

func TestParseArgsDefault(t *testing.T) {
	config, err := config.Parse()
	if err != nil {
		t.Fail()
	}

	agentConfig, err := config.AgentConfig()
	if err != nil {
		t.FailNow()
	}

	if len(agentConfig.Urls) != 1 {
		t.Fail()
	}
}
