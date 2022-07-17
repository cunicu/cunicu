package ep

import "github.com/pion/ice/v2"

type OnConnectionStateHandlerList []OnConnectionStateHandler
type OnConnectionStateHandler interface {
	OnConnectionStateChange(*Peer, ice.ConnectionState)
}

func (hl *OnConnectionStateHandlerList) Register(h OnConnectionStateHandler) {
	*hl = append(*hl, h)
}

func (hl *OnConnectionStateHandlerList) Invoke(p *Peer, cs ice.ConnectionState) {
	for _, h := range *hl {
		h.OnConnectionStateChange(p, cs)
	}
}
