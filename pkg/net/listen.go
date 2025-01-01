// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package net

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"

	"cunicu.li/cunicu/pkg/os/systemd"
)

var (
	errInvalidPortRange   = errors.New("minimal port must be larger than maximal port number")
	errInvalidNetwork     = errors.New("unsupported network")
	errNoPortFound        = errors.New("failed to find port")
	errNoSystemdListeners = errors.New("no file descriptors passed from systemd")
)

func FindRandomPortToListen(network string, mini, maxi int) (int, error) {
	if maxi < mini {
		return -1, errInvalidPortRange
	}

	if !strings.HasPrefix(network, "udp") {
		return -1, fmt.Errorf("%w: %s", errInvalidNetwork, network)
	}

	for attempts := 100; attempts > 0; attempts-- {
		port := mini + rand.Intn(maxi-mini+1) //nolint:gosec
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

func Listen(socket string) (l net.Listener, err error) {
	var network, address string

	if p := strings.SplitN(socket, ":", 2); len(p) >= 2 {
		network = p[0]
		address = p[1]
	} else if p[0] == "systemd" { //nolint:goconst
		network = p[0]
	} else {
		network = "unix"
		address = p[0]
	}

	switch {
	case network == "systemd" && address == "":
		sdListeners, err := systemd.Listeners()
		if err != nil {
			return nil, fmt.Errorf("failed to get listeners from systemd: %w", err)
		}

		if len(sdListeners) == 0 {
			return nil, errNoSystemdListeners
		}

		l = sdListeners[0]

	case network == "systemd" && address != "":
		sdListeners, err := systemd.ListenersWithNames()
		if err != nil {
			return nil, fmt.Errorf("failed to get listeners from systemd: %w", err)
		}

		if ls, ok := sdListeners[address]; !ok || len(ls) == 0 {
			return nil, fmt.Errorf("%w: with name %s", errNoSystemdListeners, address)
		} else {
			l = ls[0]
		}

	case network == "unix":
		if err := os.RemoveAll(address); err != nil {
			return nil, fmt.Errorf("failed to remove old socket: %w", err)
		}

		fallthrough

	default:
		if l, err = net.Listen(network, address); err != nil {
			return nil, fmt.Errorf("failed to listen at %s: %w", socket, err)
		}
	}

	return l, nil
}
