package ice_test

import (
	"testing"

	"github.com/pion/ice/v2"

	icex "riasc.eu/wice/internal/ice"
)

func TestCandidateTypeFromString(t *testing.T) {
	for _, ct := range []ice.CandidateType{
		ice.CandidateTypeHost,
		ice.CandidateTypeServerReflexive,
		ice.CandidateTypePeerReflexive,
		ice.CandidateTypeRelay,
	} {
		if ctp, err := icex.CandidateTypeFromString(ct.String()); err != nil || ctp != ct {
			t.Fail()
		}
	}
}

func TestNetworkTypeFromString(t *testing.T) {
	for _, nt := range []ice.NetworkType{
		ice.NetworkTypeUDP4,
		ice.NetworkTypeUDP6,
		ice.NetworkTypeTCP4,
		ice.NetworkTypeTCP6,
	} {
		if ntp, err := icex.NetworkTypeFromString(nt.String()); err != nil || ntp != nt {
			t.Fail()
		}
	}
}
