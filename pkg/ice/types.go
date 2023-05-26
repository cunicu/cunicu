// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package ice

import (
	"errors"
	"fmt"

	"github.com/pion/ice/v2"
	"github.com/pion/stun"
)

var (
	errUnknownCandidateType = errors.New("unknown candidate type")
	errUnknownNetworkType   = errors.New("unknown network type")
)

type URL struct {
	ice.URL
}

func (u *URL) UnmarshalText(text []byte) error {
	up, err := stun.ParseURI(string(text))
	if err != nil {
		return err
	}

	u.URL = *up

	return nil
}

func (u URL) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

type CandidateType struct {
	ice.CandidateType
}

func (t *CandidateType) UnmarshalText(text []byte) error {
	switch string(text) {
	case "host":
		t.CandidateType = ice.CandidateTypeHost
	case "srflx":
		t.CandidateType = ice.CandidateTypeServerReflexive
	case "prflx":
		t.CandidateType = ice.CandidateTypePeerReflexive
	case "relay":
		t.CandidateType = ice.CandidateTypeRelay
	default:
		t.CandidateType = ice.CandidateTypeUnspecified
		return fmt.Errorf("%w: %s", errUnknownCandidateType, text)
	}

	return nil
}

func (t CandidateType) MarshalText() ([]byte, error) {
	if t.CandidateType == ice.CandidateTypeUnspecified {
		return nil, ice.ErrUnknownType
	}

	return []byte(t.String()), nil
}

type NetworkType struct {
	ice.NetworkType
}

func (t *NetworkType) UnmarshalText(text []byte) error {
	switch string(text) {
	case "udp4":
		t.NetworkType = ice.NetworkTypeUDP4
	case "udp6":
		t.NetworkType = ice.NetworkTypeUDP6
	case "tcp4":
		t.NetworkType = ice.NetworkTypeTCP4
	case "tcp6":
		t.NetworkType = ice.NetworkTypeTCP6
	default:
		t.NetworkType = ice.NetworkTypeTCP4
		return fmt.Errorf("%w: %s", errUnknownNetworkType, text)
	}

	return nil
}

func (t NetworkType) MarshalText() ([]byte, error) {
	if t.String() == ice.ErrUnknownType.Error() {
		return nil, ice.ErrUnknownType
	}

	return []byte(t.String()), nil
}
