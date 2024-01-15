// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package ice

import (
	"fmt"

	"github.com/pion/ice/v3"
)

func ParseCandidateType(s string) (ice.CandidateType, error) {
	switch s {
	case "host":
		return ice.CandidateTypeHost, nil
	case "srflx":
		return ice.CandidateTypeServerReflexive, nil
	case "prflx":
		return ice.CandidateTypePeerReflexive, nil
	case "relay":
		return ice.CandidateTypeRelay, nil
	default:
		return ice.CandidateTypeUnspecified, fmt.Errorf("%w: %s", ice.ErrUnknownCandidateTyp, s)
	}
}

func ParseNetworkType(s string) (ice.NetworkType, error) {
	switch s {
	case "udp4":
		return ice.NetworkTypeUDP4, nil
	case "udp6":
		return ice.NetworkTypeUDP6, nil
	case "tcp4":
		return ice.NetworkTypeTCP4, nil
	case "tcp6":
		return ice.NetworkTypeTCP6, nil
	default:
		return 0, fmt.Errorf("%w: %s", ice.ErrUnknownType, s)
	}
}
