package hooks

import (
	"github.com/stv0g/cunicu/pkg/core"
	coreproto "github.com/stv0g/cunicu/pkg/proto/core"
)

func marshalRedactedInterface(i *core.Interface) *coreproto.Interface {
	return i.MarshalWithPeers(func(p *core.Peer) *coreproto.Peer {
		return p.Marshal().Redact()
	}).Redact()
}

func marshalRedactedPeer(p *core.Peer) *coreproto.Peer {
	return p.Marshal().Redact()
}
