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
		if err := wgCmd.Execute(); err != nil {
			os.Exit(1)
		}
	} else {
		if err := RootCmd.Execute(); err != nil {
			os.Exit(1)
		}
	}
}
