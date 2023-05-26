// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package wg

import (
	"bytes"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func CmpDevices(a, b wgtypes.Device) int {
	return bytes.Compare(a.PublicKey[:], b.PublicKey[:])
}

func CmpPeers(a, b wgtypes.Peer) int {
	return bytes.Compare(a.PublicKey[:], b.PublicKey[:])
}

func CmpPeerHandshakeTime(a, b wgtypes.Peer) int {
	if a.LastHandshakeTime.UnixMilli() == 0 && b.LastHandshakeTime.UnixMilli() != 0 {
		return 1
	}

	if b.LastHandshakeTime.UnixMilli() == 0 && a.LastHandshakeTime.UnixMilli() != 0 {
		return -1
	}

	diff := a.LastHandshakeTime.UnixMilli() - b.LastHandshakeTime.UnixMilli()
	switch {
	case diff < 0:
		return 1
	case diff > 0:
		return -1
	default:
		return 0
	}
}
