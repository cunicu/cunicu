// SPDX-FileCopyrightText: 2015 CoreOS, Inc.
// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package systemd

import (
	"net"
)

// Listeners returns a slice of net.Listener instances.
func Listeners() (listeners []net.Listener, err error) {
	files := Files(true)

	for _, f := range files {
		l, err := net.FileListener(f)
		if err != nil {
			continue
		}

		listeners = append(listeners, l)

		f.Close()
	}

	return listeners, nil
}

// ListenersWithNames maps a listener name to a set of net.Listener instances.
func ListenersWithNames() (map[string][]net.Listener, error) {
	files := Files(true)
	listeners := map[string][]net.Listener{}

	for _, f := range files {
		l, err := net.FileListener(f)
		if err != nil {
			continue
		}

		if current, ok := listeners[f.Name()]; !ok {
			listeners[f.Name()] = []net.Listener{l}
		} else {
			listeners[f.Name()] = append(current, l)
		}

		f.Close()
	}

	return listeners, nil
}
