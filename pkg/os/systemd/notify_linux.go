// SPDX-FileCopyrightText: 2014 Docker, Inc.
// SPDX-FileCopyrightText: 2015-2018 CoreOS, Inc.
// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package systemd

import (
	"net"
	"os"
	"strings"

	"cunicu.li/cunicu/pkg/log"
	"go.uber.org/zap"
)

// Notify sends a message to the init daemon. It is common to ignore the error.
// If `unsetEnv` is true, the environment variable `NOTIFY_SOCKET` will be
// unconditionally unset.
//
// It returns one of the following:
// (false, nil) - notification not supported (i.e. NOTIFY_SOCKET is unset)
// (false, err) - notification supported, but failure happened (e.g. error connecting to NOTIFY_SOCKET or while sending data)
// (true, nil) - notification supported, data has been sent
func Notify(unsetEnv bool, messages ...string) (bool, error) {
	socketAddr := &net.UnixAddr{
		Name: os.Getenv("NOTIFY_SOCKET"),
		Net:  "unixgram",
	}

	if socketAddr.Name == "" {
		return false, nil
	}

	if unsetEnv {
		if err := os.Unsetenv("NOTIFY_SOCKET"); err != nil {
			return false, err
		}
	}

	conn, err := net.DialUnix(socketAddr.Net, nil, socketAddr)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	logger := log.Global.Named("systemd")
	logger.DebugV(5, "Notifying", zap.Strings("message", messages))

	state := strings.Join(messages, "\n")

	if _, err = conn.Write([]byte(state)); err != nil {
		return false, err
	}

	return true, nil
}
