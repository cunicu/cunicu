package util

import (
	"bytes"
	"fmt"
	mrand "math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func CmpEndpoint(a, b *net.UDPAddr) int {
	if a == nil && b == nil {
		return 0
	}
	if (a != nil && b == nil) || (a == nil && b != nil) {
		return 1
	}
	if !a.IP.Equal(b.IP) || a.Port != b.Port || a.Zone != b.Zone {
		return 1
	}
	return 0
}

func CmpNet(a, b *net.IPNet) int {
	cmp := bytes.Compare(a.Mask, b.Mask)
	if cmp != 0 {
		return cmp
	}

	return bytes.Compare(a.IP, b.IP)
}

func IsATTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		panic(fmt.Errorf("failed to stat stdout: %w", err))
	}

	return (fi.Mode() & os.ModeCharDevice) != 0
}

func SetupRand() {
	mrand.Seed(time.Now().UTC().UnixNano())
}

func SetupSignals(extraSignals ...os.Signal) chan os.Signal {
	signals := []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	signals = append(signals, extraSignals...)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)

	return ch
}
