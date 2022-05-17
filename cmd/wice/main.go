package main

import (
	"os"

	"go.uber.org/zap"

	_ "riasc.eu/wice/pkg/signaling/grpc"
	_ "riasc.eu/wice/pkg/signaling/inprocess"
	_ "riasc.eu/wice/pkg/signaling/k8s"
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
