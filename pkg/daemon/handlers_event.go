// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package daemon

import (
	"net"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/stv0g/cunicu/pkg/wg"
)

type Event any

type InterfaceAddedEvent struct {
	Interface *Interface
}

type InterfaceRemovedEvent struct {
	Interface *Interface
}

type InterfaceModifiedEvent struct {
	Interface *Interface
	Old       *wg.Interface
	Modified  InterfaceModifier
}

type PeerAddedEvent struct {
	Peer *Peer
}

type PeerRemovedEvent struct {
	Peer *Peer
}

type PeerModifiedEvent struct {
	Peer              *Peer
	Old               *wgtypes.Peer
	Modified          PeerModifier
	AllowedIPsAdded   []net.IPNet
	AllowedIPsRemoved []net.IPNet
}

type EventsHandler struct {
	Events chan Event
}

func NewEventsHandler(length int) *EventsHandler {
	return &EventsHandler{
		Events: make(chan Event, length),
	}
}

func (h *EventsHandler) OnInterfaceAdded(i *Interface) {
	h.Events <- InterfaceAddedEvent{i}
}

func (h *EventsHandler) OnInterfaceRemoved(i *Interface) {
	h.Events <- InterfaceRemovedEvent{i}
}

func (h *EventsHandler) OnInterfaceModified(i *Interface, old *wg.Interface, m InterfaceModifier) {
	h.Events <- InterfaceModifiedEvent{i, old, m}
}

func (h *EventsHandler) OnPeerAdded(p *Peer) {
	h.Events <- PeerAddedEvent{p}
}

func (h *EventsHandler) OnPeerRemoved(p *Peer) {
	h.Events <- PeerRemovedEvent{p}
}

func (h *EventsHandler) OnPeerModified(p *Peer, old *wgtypes.Peer, m PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {
	h.Events <- PeerModifiedEvent{p, old, m, ipsAdded, ipsRemoved}
}
