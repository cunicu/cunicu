package ice

import (
	"fmt"

	"github.com/pion/ice/v2"
)

type URL struct {
	ice.URL
}

type CandidateType struct {
	ice.CandidateType
}

type NetworkType struct {
	ice.NetworkType
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
		return fmt.Errorf("unknown candidate type: %s", text)
	}

	return nil
}

func (t CandidateType) MarshalText() ([]byte, error) {
	if t.CandidateType == ice.CandidateTypeUnspecified {
		return nil, ice.ErrUnknownType
	}

	return []byte(t.String()), nil
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
		return fmt.Errorf("unknown network type: %s", text)
	}

	return nil
}

func (t NetworkType) MarshalText() ([]byte, error) {
	if t.String() == ice.ErrUnknownType.Error() {
		return nil, ice.ErrUnknownType
	}

	return []byte(t.String()), nil
}

func (u *URL) UnmarshalText(text []byte) error {
	up, err := ice.ParseURL(string(text))
	if err != nil {
		return err
	}

	u.URL = *up

	return nil
}

func (u URL) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

const (
	ConnectionStateCreating ConnectionState = 100 + iota
	ConnectionStateIdle
	ConnectionStateConnecting
	ConnectionStateClosing
)

type ConnectionState ice.ConnectionState

func (cs ConnectionState) String() string {
	switch cs {
	case ConnectionStateCreating:
		return "Creating"
	case ConnectionStateIdle:
		return "Idle"
	case ConnectionStateConnecting:
		return "Connecting"
	case ConnectionStateClosing:
		return "Closing"
	default:
		return ice.ConnectionState(cs).String()
	}
}
