package main_test

import (
	"flag"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

var logLevel = flag.String("log-level", "info", "log level (one of \"panic\", \"fatal\", \"error\", \"warn\", \"info\", \"debug\", \"trace\")")

func TestMain(m *testing.M) {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:  true,
		DisableQuote: true,
	})

	if lvl, err := log.ParseLevel(*logLevel); err != nil {
		log.Fatalf("invalid log level: %s", *logLevel)
	} else {
		log.SetLevel(lvl)
		log.WithField("level", *logLevel).Debug("Set log level")
	}

	rc := m.Run()
	os.Exit(rc)
}
