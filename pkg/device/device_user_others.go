// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build !windows

package device

import (
	"fmt"
	"net"

	"golang.zx2c4.com/wireguard/ipc"
)

func ListenUAPI(name string) (net.Listener, error) {
	file, err := ipc.UAPIOpen(name)
	if err != nil {
		return nil, fmt.Errorf("failed to open UAPI socket: %w", err)
	}

	return ipc.UAPIListen(name, file)
}
