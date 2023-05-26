// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package daemon

import "strings"

type PeerModifier uint32

const (
	PeerModifiedPresharedKey PeerModifier = (1 << iota)
	PeerModifiedEndpoint
	PeerModifiedKeepaliveInterval
	PeerModifiedHandshakeTime
	PeerModifiedReceiveBytes
	PeerModifiedTransmitBytes
	PeerModifiedAllowedIPs
	PeerModifiedProtocolVersion
	PeerModifiedName

	PeerModifierCount              = 8
	PeerModifiedNone  PeerModifier = 0
)

//nolint:gochecknoglobals
var PeerModifiersStrings = []string{
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

func (i PeerModifier) Strings() []string {
	modifiers := []string{}

	for j := 0; j <= PeerModifierCount; j++ {
		if i&(1<<j) != 0 {
			modifiers = append(modifiers, PeerModifiersStrings[j])
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
