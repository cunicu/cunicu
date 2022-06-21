package core

import "strings"

type PeerModifier uint32

const (
	PeerModifiedNone              PeerModifier = 0
	PeerModifiedEndpoint          PeerModifier = (1 << 0)
	PeerModifiedKeepaliveInterval PeerModifier = (1 << 1)
	PeerModifiedProtocolVersion   PeerModifier = (1 << 2)
	PeerModifiedAllowedIPs        PeerModifier = (1 << 3)
	PeerModifiedHandshakeTime     PeerModifier = (1 << 4)
)

var (
	peerModifiersStrings = []string{
		"endpoint",
		"keepalive-interval",
		"protocol-version",
		"allowed-ips",
		"handshake-time",
	}
)

func (i PeerModifier) Strings() []string {
	modifiers := []string{}

	for j := 0; j <= 4; j++ {
		if i&(1<<j) != 0 {
			modifiers = append(modifiers, peerModifiersStrings[j])
		}
	}

	return modifiers
}

func (i PeerModifier) String() string {
	return strings.Join(i.Strings(), ",")
}

func (i PeerModifier) Is(j PeerModifier) bool {
	return i&j > 0
}
