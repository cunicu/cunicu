// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package daemon

import (
	"github.com/stv0g/cunicu/pkg/crypto"
)

// InterfaceList stores all WireGuard interfaces indexed by their unique ifindex
type InterfaceList map[string]*Interface

func (l *InterfaceList) ByIndex(index int) *Interface {
	for _, i := range *l {
		if i.Index() == index {
			return i
		}
	}

	return nil
}

func (l *InterfaceList) ByName(name string) *Interface {
	return (*l)[name]
}

func (l *InterfaceList) ByPublicKey(pk crypto.Key) *Interface {
	for _, i := range *l {
		if i.PublicKey() == pk {
			return i
		}
	}

	return nil
}
