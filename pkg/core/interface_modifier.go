package core

import "strings"

type InterfaceModifier int

const (
	InterfaceModifiedNone         InterfaceModifier = 0
	InterfaceModifiedName         InterfaceModifier = (1 << 0)
	InterfaceModifiedType         InterfaceModifier = (1 << 1)
	InterfaceModifiedPrivateKey   InterfaceModifier = (1 << 2)
	InterfaceModifiedListenPort   InterfaceModifier = (1 << 3)
	InterfaceModifiedFirewallMark InterfaceModifier = (1 << 4)
	InterfaceModifiedPeers        InterfaceModifier = (1 << 5)
)

var (
	InterfaceModifiersStrings = []string{
		"name",
		"type",
		"private-key",
		"listen-port",
		"firewall-mark",
		"peers",
	}
)

func (i InterfaceModifier) Strings() []string {
	modifiers := []string{}

	for j := 0; j <= 5; j++ {
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
