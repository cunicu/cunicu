// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package daemon

import (
	"github.com/stv0g/cunicu/pkg/wg"
)

type AllHandler interface {
	InterfaceHandler
	InterfaceModifiedHandler
	PeerHandler
	PeerModifiedHandler
}

type allHandler struct {
	AllHandler
}

func (h *allHandler) OnInterfaceAdded(i *Interface) {
	i.AddModifiedHandler(h)
	i.AddPeerHandler(h)

	h.AllHandler.OnInterfaceAdded(i)
}

func (h *allHandler) OnInterfaceRemoved(i *Interface) {
	h.AllHandler.OnInterfaceRemoved(i)
}

func (h *allHandler) OnInterfaceModified(i *Interface, old *wg.Interface, m InterfaceModifier) {
	h.AllHandler.OnInterfaceModified(i, old, m)
}

func (h *allHandler) OnPeerAdded(p *Peer) {
	p.AddModifiedHandler(h)

	h.AllHandler.OnPeerAdded(p)
}

// Peer handler

type peerHandler struct {
	PeerHandler
}

func (h *peerHandler) OnInterfaceAdded(i *Interface) {
	i.AddPeerHandler(h)
}

func (h *peerHandler) OnInterfaceRemoved(_ *Interface) {}
