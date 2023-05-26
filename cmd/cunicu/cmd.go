// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//nolint:gci
package main

import (
	// Daemon features
	_ "github.com/stv0g/cunicu/pkg/daemon/feature/autocfg"
	_ "github.com/stv0g/cunicu/pkg/daemon/feature/cfgsync"
	_ "github.com/stv0g/cunicu/pkg/daemon/feature/epdisc"
	_ "github.com/stv0g/cunicu/pkg/daemon/feature/hooks"
	_ "github.com/stv0g/cunicu/pkg/daemon/feature/hsync"
	_ "github.com/stv0g/cunicu/pkg/daemon/feature/pdisc"
	_ "github.com/stv0g/cunicu/pkg/daemon/feature/rtsync"

	// Signaling backends
	_ "github.com/stv0g/cunicu/pkg/signaling/grpc"
	_ "github.com/stv0g/cunicu/pkg/signaling/inprocess"
)
