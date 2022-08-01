//go:build !windows

package util

import (
	"os"
	"os/signal"

	"golang.org/x/sys/unix"
)

const (
	SigUpdate = unix.SIGUSR1
)

func SetupSignals() chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, SigUpdate)

	return ch
}
