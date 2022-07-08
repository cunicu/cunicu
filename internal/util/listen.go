package util

import (
	"fmt"
	"math"
	"math/rand"
	"net"
	"strings"
)

func FindRandomPortToListen(network string, min, max int) (int, error) {
	if max < min {
		return -1, fmt.Errorf("minimal port must be larger than maximal port number")
	}
	if !strings.HasPrefix(network, "udp") {
		return -1, fmt.Errorf("unsupported network: %s", network)
	}

	for attempts := 100; attempts > 0; attempts-- {
		//#nosec G404 -- Port numbers do not require to be cryptographically random
		port := min + rand.Intn(max-min+1)
		if canListenOnPort(network, port) {
			return port, nil
		}
	}

	return -1, fmt.Errorf("failed to find port")
}

func FindNextPortToListen(network string, start, end int) (int, error) {
	if end == 0 {
		end = math.MaxUint16
	}

	if end < start {
		return -1, fmt.Errorf("minimal port must be larger than maximal port number")
	}
	if !strings.HasPrefix(network, "udp") {
		return -1, fmt.Errorf("unsupported network: %s", network)
	}

	for port := start; port <= end; port++ {
		if canListenOnPort(network, port) {
			return port, nil
		}
	}

	return -1, fmt.Errorf("failed to find port")
}

func canListenOnPort(network string, port int) bool {
	if conn, err := net.ListenUDP(network, &net.UDPAddr{Port: port}); err == nil {
		return conn.Close() == nil
	} else {
		return false
	}
}
