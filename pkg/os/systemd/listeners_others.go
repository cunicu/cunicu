// SPDX-FileCopyrightText: 2015 CoreOS, Inc.
// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build !linux

package systemd

import (
	"net"
)

// Listeners returns a slice of net.Listener instances.
func Listeners() (listeners []net.Listener, err error) {
	return listeners, nil
}

// ListenersWithNames maps a listener name to a set of net.Listener instances.
func ListenersWithNames() (map[string][]net.Listener, error) {
	return map[string][]net.Listener{}, nil
}
