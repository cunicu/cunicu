// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package hooks

import (
	"cunicu.li/cunicu/pkg/daemon"
	coreproto "cunicu.li/cunicu/pkg/proto/core"
)

func marshalRedactedInterface(i *daemon.Interface) *coreproto.Interface {
	return i.MarshalWithPeers(func(p *daemon.Peer) *coreproto.Peer {
		return p.Marshal().Redact()
	}).Redact()
}
