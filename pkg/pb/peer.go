package pb

import (
	"riasc.eu/wice/pkg/crypto"
)

func (p *Peer) Key() crypto.Key {
	return *(*crypto.Key)(p.PublicKey)
}
