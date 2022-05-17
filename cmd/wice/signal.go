package main

import (
	"net"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"riasc.eu/wice/internal"
	"riasc.eu/wice/pkg/signaling/grpc"
)

var (
	signalCmd = &cobra.Command{
		Use:   "signal",
		Short: "Start gRPC signaling server",

		// The main wice command is just an alias for "wice daemon"
		Run: signal,
	}

	listenAddress string
)

func init() {
	pf := signalCmd.PersistentFlags()
	pf.StringVarP(&listenAddress, "listen", "L", ":443", "listen address")

	RootCmd.AddCommand(signalCmd)
}

func signal(cmd *cobra.Command, args []string) {
	l, err := net.Listen("tcp", listenAddress)
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	svr := grpc.NewServer()

	sigs := internal.SetupSignals()
	go func() {
		for sig := range sigs {
			switch sig {
			default:
				svr.Stop()
			}
		}
	}()

	logger.Info("Starting gRPC signaling server", zap.String("address", listenAddress))

	if err := svr.Serve(l); err != nil {
		logger.Fatal("Failed to start gRPC server", zap.Error(err))
	}
}
