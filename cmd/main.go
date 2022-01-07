package main

import (
	"go.uber.org/zap"

	_ "riasc.eu/wice/pkg/signaling/k8s"
	_ "riasc.eu/wice/pkg/signaling/p2p"
)

var (
	logger *zap.Logger
)

func main() {
	rootCmd.Execute()
}
