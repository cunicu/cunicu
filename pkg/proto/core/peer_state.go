// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"strings"

	"github.com/stv0g/cunicu/pkg/tty"
)

func (s PeerState) Color() string {
	switch s {
	case PeerState_CONNECTING:
		return tty.FgYellow
	case PeerState_CONNECTED:
		return tty.FgGreen
	case PeerState_FAILED:
		return tty.FgRed
	case PeerState_NEW, PeerState_CLOSED:
		return tty.FgWhite
	}

	return tty.FgDefault
}

func (s *PeerState) MarshalText() ([]byte, error) {
	return []byte(strings.ToLower(s.String())), nil
}
