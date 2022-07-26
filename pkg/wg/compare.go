package wg

import (
	"bytes"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func CmpDevices(a, b *wgtypes.Device) int {
	return bytes.Compare(a.PublicKey[:], b.PublicKey[:])
}

func CmpPeers(a, b *wgtypes.Peer) int {
	return bytes.Compare(a.PublicKey[:], b.PublicKey[:])
}

func CmpPeerHandshakeTime(a, b *wgtypes.Peer) int {
	if a.LastHandshakeTime.UnixMilli() == 0 && b.LastHandshakeTime.UnixMilli() != 0 {
		return 1
	}

	if b.LastHandshakeTime.UnixMilli() == 0 && a.LastHandshakeTime.UnixMilli() != 0 {
		return -1
	}

	diff := a.LastHandshakeTime.UnixMilli() - b.LastHandshakeTime.UnixMilli()
	if diff < 0 {
		return 1
	} else if diff > 0 {
		return -1
	} else {
		return 0
	}
}
