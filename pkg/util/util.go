package util

import (
	"bytes"
	crand "crypto/rand"
	"encoding/base64"
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

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := crand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

func IsATTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		panic(fmt.Errorf("failed to stat stdout: %w", err))
	}

	return (fi.Mode() & os.ModeCharDevice) != 0
}

func LastTime(ts ...time.Time) time.Time {
	var lt time.Time

	for _, t := range ts {
		if t.After(lt) {
			lt = t
		}
	}

	return lt
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
