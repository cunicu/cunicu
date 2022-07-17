package watcher

import "riasc.eu/wice/pkg/core"

type handler struct {
	core.AllHandler
}

func (h *handler) OnInterfaceAdded(i *core.Interface) {
	i.OnPeer(h)
	i.OnModified(h)

	h.AllHandler.OnInterfaceAdded(i)
}

func (h *handler) OnPeerAdded(p *core.Peer) {
	p.OnModified(h)

	h.AllHandler.OnPeerAdded(p)
}

// RegisterAll adds a new handler to all the events observed by the watcher.
func (w *Watcher) RegisterAll(h core.AllHandler) {
	w.OnInterface(&handler{h})
}
