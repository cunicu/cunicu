package main

import (
	"os"

	"go.uber.org/zap"

	_ "riasc.eu/wice/pkg/signaling/grpc"
	_ "riasc.eu/wice/pkg/signaling/k8s"
	_ "riasc.eu/wice/pkg/signaling/p2p"
)

var (
	logger *zap.Logger
)

func main() {
	if os.Args[0] == "wg" {
		wgCmd.Execute()
	} else {
		rootCmd.Execute()
	}
}
