// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package daemon

import (
	"golang.org/x/exp/slices"

	"github.com/stv0g/cunicu/pkg/wg"
)

type InterfaceModifiedHandler interface {
	OnInterfaceModified(i *Interface, old *wg.Interface, m InterfaceModifier)
}

type PeerStateChangedHandler interface {
	OnPeerStateChanged(p *Peer, newState, prevState PeerState)
}

type PeerHandler interface {
	OnPeerAdded(p *Peer)
	OnPeerRemoved(p *Peer)
}

func (i *Interface) AddPeerStateChangeHandler(h PeerStateChangedHandler) {
	if !slices.Contains(i.onPeerStateChanged, h) {
		i.onPeerStateChanged = append(i.onPeerStateChanged, h)
	}
}

func (i *Interface) RemovePeerStateChangeHandler(h PeerStateChangedHandler) {
	if idx := slices.Index(i.onPeerStateChanged, h); idx > -1 {
		i.onPeerStateChanged = slices.Delete(i.onPeerStateChanged, idx, idx+1)
	}
}

func (i *Interface) AddModifiedHandler(h InterfaceModifiedHandler) {
	if !slices.Contains(i.onModified, h) {
		i.onModified = append(i.onModified, h)
	}
}

func (i *Interface) RemoveModifiedHandler(h InterfaceModifiedHandler) {
	if idx := slices.Index(i.onModified, h); idx > -1 {
		i.onModified = slices.Delete(i.onModified, idx, idx+1)
	}
}

func (i *Interface) AddPeerHandler(h PeerHandler) {
	if !slices.Contains(i.onPeer, h) {
		i.onPeer = append(i.onPeer, h)
	}
}

func (i *Interface) RemovePeerHandler(h PeerHandler) {
	if idx := slices.Index(i.onPeer, h); idx > -1 {
		i.onPeer = slices.Delete(i.onPeer, idx, idx+1)
	}
}
