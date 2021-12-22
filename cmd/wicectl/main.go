package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"riasc.eu/wice/pkg/socket"
)

var sockPath = flag.String("socket", "/var/run/wice.go", "Unix control and monitoring socket")

func setupLogging() {
	log.SetFormatter(&log.TextFormatter{
		// ForceColors:  true,
		// DisableQuote: true,
	})
}

func setupSignals() chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	return ch
}

func main() {
	setupLogging()
	signals := setupSignals()

	flag.Parse()

	subCommand := flag.Arg(0)

	switch subCommand {
	case "monitor":
		monitor(signals)

	default:
		log.Fatalf("Unknown subcommand: %s", subCommand)
	}
}

func monitor(signals chan os.Signal) {
	logger := log.WithField("logger", "events")

	sock, err := socket.Connect(*sockPath)
	if err != nil {
		log.WithError(err).Fatalf("Failed to connect to control socket: %s", err)
	}

	for {
		select {
		case sig := <-signals:
			log.Info("Received signal: %s", sig)
			os.Exit(0)

		case evt := <-sock.Events:
			evt.Log(logger, "Event")
		}
	}
}
