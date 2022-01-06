package config

import (
	"fmt"

	"github.com/pion/ice/v2"
)

func candidateTypeFromString(t string) (ice.CandidateType, error) {
	switch t {
	case "host":
		return ice.CandidateTypeHost, nil
	case "srflx":
		return ice.CandidateTypeServerReflexive, nil
	case "prflx":
		return ice.CandidateTypePeerReflexive, nil
	case "relay":
		return ice.CandidateTypeRelay, nil
	default:
		return ice.CandidateTypeUnspecified, fmt.Errorf("unknown candidate type: %s", t)
	}
}

func networkTypeFromString(t string) (ice.NetworkType, error) {
	switch t {
	case "udp4":
		return ice.NetworkTypeUDP4, nil
	case "udp6":
		return ice.NetworkTypeUDP6, nil
	case "tcp4":
		return ice.NetworkTypeTCP4, nil
	case "tcp6":
		return ice.NetworkTypeTCP6, nil
	default:
		return ice.NetworkTypeTCP4, fmt.Errorf("unknown network type: %s", t)
	}
}
