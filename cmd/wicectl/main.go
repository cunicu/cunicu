package main

import (
	"flag"
	"os"

	"go.uber.org/zap"
	"riasc.eu/wice/internal"
	"riasc.eu/wice/pkg/socket"
)

var sockPath = flag.String("socket", "/var/run/wice.go", "Unix control and monitoring socket")

func main() {
	internal.SetupRand()
	signals := internal.SetupSignals()
	logger := internal.SetupLogging()
	defer logger.Sync()

	flag.Parse()

	subCommand := flag.Arg(0)

	switch subCommand {
	case "monitor":
		monitor(signals)

	default:
		logger.Fatal("Unknown subcommand", zap.String("command", subCommand))
	}
}

func monitor(signals chan os.Signal) {
	logger := zap.L().Named("events")

	sock, err := socket.Connect(*sockPath)
	if err != nil {
		logger.Fatal("Failed to connect to control socket", zap.Error(err))
	}

	for {
		select {
		case sig := <-signals:
			logger.Info("Received signal", zap.Any("signal", sig))
			os.Exit(0)

		case evt := <-sock.Events:
			evt.Log(logger, "Event")
		}
	}
}
