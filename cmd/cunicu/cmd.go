// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//nolint:gci
package main

import (
	// Daemon features
	_ "cunicu.li/cunicu/pkg/daemon/feature/autocfg"
	_ "cunicu.li/cunicu/pkg/daemon/feature/cfgsync"
	_ "cunicu.li/cunicu/pkg/daemon/feature/epdisc"
	_ "cunicu.li/cunicu/pkg/daemon/feature/hooks"
	_ "cunicu.li/cunicu/pkg/daemon/feature/hsync"
	_ "cunicu.li/cunicu/pkg/daemon/feature/pdisc"
	_ "cunicu.li/cunicu/pkg/daemon/feature/rtsync"

	// Signaling backends
	_ "cunicu.li/cunicu/pkg/signaling/grpc"
	_ "cunicu.li/cunicu/pkg/signaling/inprocess"
)
