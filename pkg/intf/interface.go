package intf

import (
	"io"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Interface interface {
	io.Closer

	DumpConfig(io.Writer)

	AddPeer(peer wgtypes.Key) error
	RemovePeer(peer wgtypes.Key) error

	Sync(*wgtypes.Device) error

	Name() string
}
