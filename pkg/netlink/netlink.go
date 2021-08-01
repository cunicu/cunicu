package netlink

import "github.com/vishvananda/netlink"

const (
	LinkTypeWireguard = "wireguard"
)

// github.com/vishvananda/netlink does not come with the wireguard link type yet
type Wireguard struct {
	netlink.LinkAttrs
}

func (wg *Wireguard) Attrs() *netlink.LinkAttrs {
	return &wg.LinkAttrs
}

func (wg *Wireguard) Type() string {
	return LinkTypeWireguard
}
