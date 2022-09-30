package core

import (
	"net"

	"github.com/stv0g/cunicu/pkg/wg"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type InterfaceHandler interface {
	OnInterfaceAdded(i *Interface)
	OnInterfaceRemoved(i *Interface)
}

type InterfaceModifiedHandler interface {
	OnInterfaceModified(i *Interface, old *wg.Device, m InterfaceModifier)
}

type PeerHandler interface {
	OnPeerAdded(p *Peer)
	OnPeerRemoved(p *Peer)
}

type PeerModifiedHandler interface {
	OnPeerModified(p *Peer, old *wgtypes.Peer, m PeerModifier, ipsAdded, ipsRemoved []net.IPNet)
}

type AllHandler interface {
	InterfaceHandler
	InterfaceModifiedHandler
	PeerHandler
	PeerModifiedHandler
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
	Peer              *Peer
	Old               *wgtypes.Peer
	Modified          PeerModifier
	AllowedIPsAdded   []net.IPNet
	AllowedIPsRemoved []net.IPNet
}

type EventsHandler struct {
	Events chan Event
}

func NewEventsHandler(len int) *EventsHandler {
	return &EventsHandler{
		Events: make(chan Event, len),
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
	h.Events <- PeerModifiedEvent{p, old, m, ipsAdded, ipsRemoved}
}
