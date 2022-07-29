package watcher

import (
	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/wg"
)

// All handler

type allHandler struct {
	core.AllHandler
}

func (h *allHandler) OnInterfaceAdded(i *core.Interface) {
	i.OnModified(h)
	i.OnPeer(h)

	h.AllHandler.OnInterfaceAdded(i)
}

func (h *allHandler) OnInterfaceRemoved(i *core.Interface) {}

func (h *allHandler) OnInterfaceModified(i *core.Interface, old *wg.Device, m core.InterfaceModifier) {
}

// Peer handler

type peerHandler struct {
	core.PeerHandler
}

func (h *peerHandler) OnInterfaceAdded(i *core.Interface) {
	i.OnPeer(h)
}

func (h *peerHandler) OnInterfaceRemoved(i *core.Interface) {}

func (h *peerHandler) OnInterfaceModified(i *core.Interface, old *wg.Device, m core.InterfaceModifier) {
}

func (h *peerHandler) OnPeerAdded(p *core.Peer) {
	p.OnModified(h)

	h.PeerHandler.OnPeerAdded(p)
}

// OnAll adds a new handler to all the events observed by the watcher.
func (w *Watcher) OnAll(h core.AllHandler) {
	w.OnInterface(&allHandler{h})
}

// OnPeer registers an handler for peer-related events
func (w *Watcher) OnPeer(h core.PeerHandler) {
	w.OnInterface(&peerHandler{h})
}

// OnInterface registers an handler for interface-related events
func (w *Watcher) OnInterface(h core.InterfaceHandler) {
	w.onInterface = append(w.onInterface, h)
}
