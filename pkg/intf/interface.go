package intf

import (
	"io"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
)

type Interface interface {
	io.Closer

	Dump(wr io.Writer, color bool, hideKeys bool)
	DumpConfig(wr io.Writer)
	SyncConfig(cfg string) error

	AddPeer(peer wgtypes.Key) error
	RemovePeer(peer wgtypes.Key) error

	Sync(*wgtypes.Device) error

	Marshal() *pb.Interface

	// Getter
	Name() string
	PublicKey() crypto.Key
	PrivateKey() crypto.Key
	Peers() map[crypto.Key]*Peer
}
