package internal

import (
	"math/rand"
	"os"
	"os/signal"
	"time"

	"golang.org/x/sys/unix"
)

func SetupRand() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func SetupSignals() chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, unix.SIGINT, unix.SIGTERM)

	return ch
}
