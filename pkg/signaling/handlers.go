package signaling

// Backend ready

type BackendReadyHandlerList []BackendReadyHandler
type BackendReadyHandler interface {
	OnBackendReady(b Backend)
}

func (hl *BackendReadyHandlerList) Register(h BackendReadyHandler) {
	*hl = append(*hl, h)
}

func (hl *BackendReadyHandlerList) Invoke(b Backend) {
	for _, h := range *hl {
		h.OnBackendReady(b)
	}
}
