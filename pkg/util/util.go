// Package util implements project-wide universal utilities
package util

import (
	"bytes"
	"encoding/binary"
	"fmt"
	mrand "math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/exp/slices"
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

func ContainsNet(outer, inner *net.IPNet) bool {
	outerOnes, _ := outer.Mask.Size()
	innerOnes, _ := inner.Mask.Size()
	return outerOnes <= innerOnes && outer.Contains(inner.IP)
}

func IsATTY(f *os.File) bool {
	fi, err := f.Stat()
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

func OffsetIP(ip net.IP, off int) net.IP {
	oip := slices.Clone(ip)

	if isV6 := ip.To4() == nil; isV6 {
		num := binary.BigEndian.Uint64(ip[8:])
		binary.BigEndian.PutUint64(oip[8:], num+uint64(off))
	} else {
		num := binary.BigEndian.Uint32(ip[12:])
		binary.BigEndian.PutUint32(oip[12:], num+uint32(off))
	}

	return oip
}
