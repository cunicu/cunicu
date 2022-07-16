package watcher

import "riasc.eu/wice/pkg/core"

type handler struct {
	core.AllHandler
}

func (h *handler) OnInterfaceAdded(i *core.Interface) {
	i.OnPeer.Register(h)
	i.OnModified.Register(h)

	h.AllHandler.OnInterfaceAdded(i)
}

func (h *handler) OnPeerAdded(p *core.Peer) {
	p.OnModified.Register(h)

	h.AllHandler.OnPeerAdded(p)
}

// RegisterAll adds a new handler to all the events observed by the watcher.
func (w *Watcher) RegisterAll(h core.AllHandler) {
	w.OnInterface.Register(&handler{h})
}
