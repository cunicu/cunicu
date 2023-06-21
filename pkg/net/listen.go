// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package net

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"strings"
)

var (
	errInvalidPortRange = errors.New("minimal port must be larger than maximal port number")
	errInvalidNetwork   = errors.New("unsupported network")
	errNoPortFound      = errors.New("failed to find port")
)

func FindRandomPortToListen(network string, min, max int) (int, error) {
	if max < min {
		return -1, errInvalidPortRange
	}
	if !strings.HasPrefix(network, "udp") {
		return -1, fmt.Errorf("%w: %s", errInvalidNetwork, network)
	}

	for attempts := 100; attempts > 0; attempts-- {
		port := min + rand.Intn(max-min+1) //nolint:gosec
		if canListenOnPort(network, port) {
			return port, nil
		}
	}

	return -1, errNoPortFound
}

func FindNextPortToListen(network string, start, end int) (int, error) {
	if end < start {
		return -1, errInvalidPortRange
	}
	if !strings.HasPrefix(network, "udp") {
		return -1, fmt.Errorf("%w: %s", errInvalidNetwork, network)
	}

	for port := start; port <= end; port++ {
		if canListenOnPort(network, port) {
			return port, nil
		}
	}

	return -1, errNoPortFound
}

func canListenOnPort(network string, port int) bool {
	if conn, err := net.ListenUDP(network, &net.UDPAddr{Port: port}); err == nil {
		return conn.Close() == nil
	}

	return false
}
