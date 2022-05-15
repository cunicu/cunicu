package util

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"strings"
)

func FindRandomPortToListen(network string, min, max int) (int, error) {
	if max < min {
		return -1, fmt.Errorf("minimal port must be larger than maximal port number")
	}

	for attempts := 100; attempts > 0; attempts-- {
		port := min + rand.Intn(max-min+1)
		if canListenOnPort(network, port) {
			return port, nil
		}
	}

	return -1, fmt.Errorf("failed to find port")
}

func FindNextPortToListen(network string, start, end int) (int, error) {
	if end < start {
		return -1, fmt.Errorf("minimal port must be larger than maximal port number")
	}

	for port := start; start <= end; port++ {
		if canListenOnPort(network, port) {
			return port, nil
		}
	}

	return -1, fmt.Errorf("failed to find port")
}

func canListenOnPort(network string, port int) bool {
	var addr = fmt.Sprintf(":%d", port)
	var conn io.Closer
	var err error

	if strings.HasPrefix(network, "udp") {
		conn, err = net.ListenPacket(network, addr)
	} else if strings.HasPrefix(network, "tcp") {
		conn, err = net.Listen(network, addr)
	}
	if err == nil {
		conn.Close()
		return true
	}

	return false
}
