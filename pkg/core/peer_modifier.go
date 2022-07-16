package core

import "strings"

type PeerModifier uint32

const (
	PeerModifiedNone              PeerModifier = 0
	PeerModifiedPresharedKey      PeerModifier = (1 << 0)
	PeerModifiedEndpoint          PeerModifier = (1 << 1)
	PeerModifiedKeepaliveInterval PeerModifier = (1 << 2)
	PeerModifiedHandshakeTime     PeerModifier = (1 << 3)
	PeerModifiedReceiveBytes      PeerModifier = (1 << 4)
	PeerModifiedTransmitBytes     PeerModifier = (1 << 5)
	PeerModifiedAllowedIPs        PeerModifier = (1 << 6)
	PeerModifiedProtocolVersion   PeerModifier = (1 << 7)
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
	}
)

func (i PeerModifier) Strings() []string {
	modifiers := []string{}

	for j := 0; j <= 7; j++ {
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
