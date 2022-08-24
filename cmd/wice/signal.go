package main

import (
	"net"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpcx "riasc.eu/wice/pkg/signaling/grpc"
	"riasc.eu/wice/pkg/util"
)

var (
	signalCmd = &cobra.Command{
		Use:   "signal",
		Short: "Start gRPC signaling server",

		Run: signal,
	}

	listenAddress string
	secure        = false
)

func init() {
	pf := signalCmd.PersistentFlags()
	pf.StringVarP(&listenAddress, "listen", "L", ":8080", "listen address")
	pf.BoolVarP(&secure, "secure", "S", false, "listen with TLS")

	rootCmd.AddCommand(signalCmd)
}

func signal(cmd *cobra.Command, args []string) {
	l, err := net.Listen("tcp", listenAddress)
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	// Disable TLS
	opts := []grpc.ServerOption{}
	if !secure {
		opts = append(opts, grpc.Creds(insecure.NewCredentials()))
	}

	svr := grpcx.NewServer(opts...)

	go func() {
		for sig := range util.SetupSignals() {
			logger.Debug("Received signal", zap.Any("signal", sig))

			svr.Close()
		}
	}()

	logger.Info("Starting gRPC signaling server", zap.String("address", listenAddress))

	if err := svr.Serve(l); err != nil {
		logger.Fatal("Failed to start gRPC server", zap.Error(err))
	}

	logger.Info("Gracefully stopped gRPC signaling server")
}
