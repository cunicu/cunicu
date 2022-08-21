package cmd

import (
	"go.uber.org/zap"
	_ "riasc.eu/wice/pkg/signaling/grpc"
	_ "riasc.eu/wice/pkg/signaling/inprocess"
	_ "riasc.eu/wice/pkg/signaling/k8s"
)

var (
	logger *zap.Logger
)
