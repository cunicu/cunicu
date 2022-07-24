package pb

import (
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func NewInterface(i *wgtypes.Device) *Interface {
	peers := []*Peer{}
	for _, peer := range i.Peers {
		peers = append(peers, NewPeer(peer))
	}

	return &Interface{
		Name:         i.Name,
		Type:         Interface_Type(i.Type),
		PrivateKey:   i.PrivateKey[:],
		PublicKey:    i.PublicKey[:],
		ListenPort:   uint32(i.ListenPort),
		FirewallMark: uint32(i.FirewallMark),
		Peers:        peers,
	}
}

func (i *Interface) Device() *wgtypes.Device {
	peers := []wgtypes.Peer{}
	for _, peer := range i.Peers {
		peers = append(peers, peer.Peer())
	}

	return &wgtypes.Device{
		Name:         i.Name,
		Type:         wgtypes.DeviceType(i.Type),
		PublicKey:    *(*wgtypes.Key)(i.PublicKey),
		PrivateKey:   *(*wgtypes.Key)(i.PrivateKey),
		ListenPort:   int(i.ListenPort),
		FirewallMark: int(i.FirewallMark),
		Peers:        peers,
	}
}
