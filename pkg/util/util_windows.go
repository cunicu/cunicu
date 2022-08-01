//go:build windows

package util

import (
	"os"
	"os/signal"
	"syscall"
)

const (
	SigUpdate = syscall.Signal(-1) // not supported
)

func SetupSignals() chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	return ch
}
