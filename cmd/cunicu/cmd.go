package main

import (
	// Signaling backends
	_ "github.com/stv0g/cunicu/pkg/signaling/grpc"
	_ "github.com/stv0g/cunicu/pkg/signaling/inprocess"
	_ "github.com/stv0g/cunicu/pkg/signaling/k8s"

	// Daemon features
	_ "github.com/stv0g/cunicu/pkg/daemon/feature/autocfg"
	_ "github.com/stv0g/cunicu/pkg/daemon/feature/cfgsync"
	_ "github.com/stv0g/cunicu/pkg/daemon/feature/epdisc"
	_ "github.com/stv0g/cunicu/pkg/daemon/feature/hooks"
	_ "github.com/stv0g/cunicu/pkg/daemon/feature/hsync"
	_ "github.com/stv0g/cunicu/pkg/daemon/feature/pdisc"
	_ "github.com/stv0g/cunicu/pkg/daemon/feature/pske"
	_ "github.com/stv0g/cunicu/pkg/daemon/feature/rtsync"
)
