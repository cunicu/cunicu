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

func SetupSignals(extraSignals ...os.Signal) chan os.Signal {
	signals := []os.Signal{unix.SIGINT, unix.SIGTERM}
	signals = append(signals, extraSignals...)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)

	return ch
}
