// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package daemon

import "strings"

type InterfaceModifier int

const (
	InterfaceModifiedName InterfaceModifier = (1 << iota)
	InterfaceModifiedType
	InterfaceModifiedPrivateKey
	InterfaceModifiedListenPort
	InterfaceModifiedFirewallMark
	InterfaceModifiedPeers

	InterfaceModifierCount                   = 6
	InterfaceModifiedNone  InterfaceModifier = 0
)

//nolint:gochecknoglobals
var InterfaceModifiersStrings = []string{
	"name",
	"type",
	"private-key",
	"listen-port",
	"firewall-mark",
	"peers",
}

func (i InterfaceModifier) Strings() []string {
	modifiers := []string{}

	for j := 0; j < InterfaceModifierCount; j++ {
		if i&(1<<j) != 0 {
			modifiers = append(modifiers, InterfaceModifiersStrings[j])
		}
	}

	return modifiers
}

func (i InterfaceModifier) String() string {
	return strings.Join(i.Strings(), ",")
}

func (i InterfaceModifier) Is(j InterfaceModifier) bool {
	return i&j > 0
}
