package main

import (
	"net"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/stv0g/cunicu/pkg/util"

	grpcx "github.com/stv0g/cunicu/pkg/signaling/grpc"
)

var (
	relayCmd = &cobra.Command{
		Use:   "relay URL...",
		Short: "Start relay API server",
		Long: `This command starts a gRPC server providing cunicu agents with a list of available STUN and TURN servers.

**Note:** Currently this command does not run a TURN server itself. But relies on an external server like Coturn.

With this feature you can distribute a list of available STUN/TURN servers easily to a fleet of agents.
It also allows to issue short-lived HMAC-SHA1 credentials based the proposed TURN REST API and thereby static long term credentials.

The command expects a list of STUN or TURN URLs according to RFC7065/RFC7064 with a few extensions:

- A secret for the TURN REST API can be provided by the 'secret' query parameter
  - Example: turn:server.com?secret=rest-api-secret

- A time-to-live to the TURN REST API secrets can be provided by the 'ttl' query parameter
  - Example: turn:server.com?ttl=1h

- Static TURN credentials can be provided by the URIs user info
  - Example: turn:user1:pass1@server.com
`,
		Example: `relay turn:server.com?secret=rest-api-secret&ttl=1h`,

		Run: relay,
	}
)

func init() {
	pf := relayCmd.PersistentFlags()
	pf.StringVarP(&listenAddress, "listen", "L", ":8080", "listen address")
	pf.BoolVarP(&secure, "secure", "S", false, "listen with TLS")

	rootCmd.AddCommand(relayCmd)
}

func relay(cmd *cobra.Command, args []string) {
	l, err := net.Listen("tcp", listenAddress)
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	// Disable TLS
	opts := []grpc.ServerOption{}
	if !secure {
		opts = append(opts, grpc.Creds(insecure.NewCredentials()))
	}

	svr, err := grpcx.NewRelayAPIServer(args, opts...)
	if err != nil {
		logger.Fatal("Failed to start gRPC server", zap.Error(err))
	}

	go func() {
		for sig := range util.SetupSignals() {
			logger.Debug("Received signal", zap.Any("signal", sig))

			if err := svr.Close(); err != nil {
				logger.Error("Failed to close server", zap.Error(err))
			}
		}
	}()

	logger.Info("Starting gRPC relay API server", zap.String("address", listenAddress))

	if err := svr.Serve(l); err != nil {
		logger.Fatal("Failed to start gRPC server", zap.Error(err))
	}

	logger.Info("Gracefully stopped gRPC relay API server")
}
