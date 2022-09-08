package main

import (
	_ "github.com/stv0g/cunicu/pkg/signaling/grpc"
	_ "github.com/stv0g/cunicu/pkg/signaling/inprocess"
	_ "github.com/stv0g/cunicu/pkg/signaling/k8s"
	"go.uber.org/zap"
)

var (
	logger *zap.Logger
)
