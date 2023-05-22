package daemon

import "golang.org/x/exp/slices"

type InterfaceHandler interface {
	OnInterfaceAdded(i *Interface)
	OnInterfaceRemoved(i *Interface)
}

// AddAllHandler adds a new handler to all the events observed by the watcher.
func (w *Watcher) AddAllHandler(h AllHandler) {
	w.AddInterfaceHandler(&allHandler{h})
}

// AddPeerHandler registers an handler for peer-related events
func (w *Watcher) AddPeerHandler(h PeerHandler) {
	w.AddInterfaceHandler(&peerHandler{h})
}

// AddInterfaceHandler registers an handler for interface-related events
func (w *Watcher) AddInterfaceHandler(h InterfaceHandler) {
	if !slices.Contains(w.onInterface, h) {
		w.onInterface = append(w.onInterface, h)
	}
}
