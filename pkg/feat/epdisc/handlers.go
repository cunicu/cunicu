package epdisc

import icex "riasc.eu/wice/pkg/feat/epdisc/ice"

type OnConnectionStateHandler interface {
	OnConnectionStateChange(p *Peer, new, prev icex.ConnectionState)
}

func (e *EndpointDiscovery) OnConnectionStateChange(h OnConnectionStateHandler) {
	e.onConnectionStateChange = append(e.onConnectionStateChange, h)
}
