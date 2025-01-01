// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package rpc

import (
	"fmt"

	rpcproto "cunicu.li/cunicu/pkg/proto/rpc"
)

type Event struct {
	rpcproto.Event
}

func (e *Event) String() string {
	return fmt.Sprintf("type=%s", e.Type)
}
