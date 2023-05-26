// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package hooks

import (
	"github.com/stv0g/cunicu/pkg/daemon"
	coreproto "github.com/stv0g/cunicu/pkg/proto/core"
)

func marshalRedactedInterface(i *daemon.Interface) *coreproto.Interface {
	return i.MarshalWithPeers(func(p *daemon.Peer) *coreproto.Peer {
		return p.Marshal().Redact()
	}).Redact()
}
