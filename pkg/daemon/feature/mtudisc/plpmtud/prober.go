package plpmtud

type Prober interface {
	SendProbeRequest(mtu uint) error
	SendProbeResponse(mtu uint) error

	RegisterDiscoverer(h *Discoverer)

	Close() error
}
