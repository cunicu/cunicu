package core

import (
	"net"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/pkg/wg"
)

type InterfaceHandler interface {
	OnInterfaceAdded(i *Interface)
	OnInterfaceRemoved(i *Interface)
	OnInterfaceModified(i *Interface, old *wg.Device, m InterfaceModifier)
}

type PeerHandler interface {
	OnPeerAdded(p *Peer)
	OnPeerRemoved(p *Peer)
	OnPeerModified(p *Peer, old *wgtypes.Peer, m PeerModifier, ipsAdded, ipsRemoved []net.IPNet)
}

type AllHandler interface {
	InterfaceHandler
	PeerHandler
}

type Event any

type InterfaceAddedEvent struct {
	Interface *Interface
}

type InterfaceRemovedEvent struct {
	Interface *Interface
}

type InterfaceModifiedEvent struct {
	Interface *Interface
	Old       *wg.Device
	Modified  InterfaceModifier
}

type PeerAddedEvent struct {
	Peer *Peer
}

type PeerRemovedEvent struct {
	Peer *Peer
}

type PeerModifiedEvent struct {
	Peer     *Peer
	Old      *wgtypes.Peer
	Modified PeerModifier
}

type EventsHandler struct {
	Events chan Event
}

func NewMockHandler() *EventsHandler {
	return &EventsHandler{
		Events: make(chan Event),
	}
}

func (h *EventsHandler) OnInterfaceAdded(i *Interface) {
	h.Events <- InterfaceAddedEvent{i}
}

func (h *EventsHandler) OnInterfaceRemoved(i *Interface) {
	h.Events <- InterfaceRemovedEvent{i}
}

func (h *EventsHandler) OnInterfaceModified(i *Interface, old *wg.Device, m InterfaceModifier) {
	h.Events <- InterfaceModifiedEvent{i, old, m}
}

func (h *EventsHandler) OnPeerAdded(p *Peer) {
	h.Events <- PeerAddedEvent{p}
}

func (h *EventsHandler) OnPeerRemoved(p *Peer) {
	h.Events <- PeerRemovedEvent{p}
}

func (h *EventsHandler) OnPeerModified(p *Peer, old *wgtypes.Peer, m PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {
	h.Events <- PeerModifiedEvent{p, old, m}
}
