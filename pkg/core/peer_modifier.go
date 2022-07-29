package core

import "strings"

type PeerModifier uint32

const (
	PeerModifiedPresharedKey      PeerModifier = (1 << iota)
	PeerModifiedEndpoint          PeerModifier = (1 << iota)
	PeerModifiedKeepaliveInterval PeerModifier = (1 << iota)
	PeerModifiedHandshakeTime     PeerModifier = (1 << iota)
	PeerModifiedReceiveBytes      PeerModifier = (1 << iota)
	PeerModifiedTransmitBytes     PeerModifier = (1 << iota)
	PeerModifiedAllowedIPs        PeerModifier = (1 << iota)
	PeerModifiedProtocolVersion   PeerModifier = (1 << iota)
	PeerModifiedName              PeerModifier = (1 << iota)
	PeerModifierCount                          = iota

	PeerModifiedNone PeerModifier = 0
)

var (
	peerModifiersStrings = []string{
		"preshared-key",
		"endpoint",
		"keepalive-interval",
		"handshake-time",
		"receive-bytes",
		"transmit-bytes",
		"allowed-ips",
		"protocol-version",
		"name",
	}
)

func (i PeerModifier) Strings() []string {
	modifiers := []string{}

	for j := 0; j <= PeerModifierCount; j++ {
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
