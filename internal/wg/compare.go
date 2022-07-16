package wg

import (
	"bytes"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Less func(i, j int) bool

func CmpDevices(a, b *wgtypes.Device) int {
	return bytes.Compare(a.PublicKey[:], b.PublicKey[:])
}

func CmpPeers(a, b *wgtypes.Peer) int {
	return bytes.Compare(a.PublicKey[:], b.PublicKey[:])
}

func LessPeers(peers []wgtypes.Peer) Less {
	return func(i, j int) bool { return CmpPeers(&peers[i], &peers[j]) < 0 }
}
