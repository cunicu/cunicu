package epice

import icex "riasc.eu/wice/pkg/ice"

type OnConnectionStateHandler interface {
	OnConnectionStateChange(p *Peer, new, prev icex.ConnectionState)
}
