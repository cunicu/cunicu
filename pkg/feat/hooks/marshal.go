package hooks

import (
	"riasc.eu/wice/pkg/core"
	coreproto "riasc.eu/wice/pkg/proto/core"
)

func marshalRedactedInterface(i *core.Interface) *coreproto.Interface {
	return i.MarshalWithPeers(func(p *core.Peer) *coreproto.Peer {
		return p.Marshal().Redact()
	}).Redact()
}

func marshalRedactedPeer(p *core.Peer) *coreproto.Peer {
	return p.Marshal().Redact()
}
