package util

import (
	"bytes"
	"encoding/base64"
	"math/rand"
	"net"
	"os"
	"syscall"
	"unsafe"
)

type Less func(i, j int) bool

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

// func lessNets(nets []net.IPNet) Less {
// 	return func(i, j int) bool { return cmpNet(&nets[i], &nets[j]) < 0 }
// }

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
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

func SetsockoptBytes(fd int, level int, opt int, b []byte) syscall.Errno {
	_, _, errno := syscall.Syscall6(syscall.SYS_SETSOCKOPT,
		uintptr(fd), uintptr(level), uintptr(opt),
		uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)), 0)

	return errno
}

func IsATTY() bool {
	fi, _ := os.Stdout.Stat()

	return (fi.Mode() & os.ModeCharDevice) != 0
}
