package netlink_test

import (
	"os"
	"testing"

	"github.com/vishvananda/netlink"
	nl "riasc.eu/wice/pkg/netlink"
)

func TestWireguardLink(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip()
	}

	l := &nl.Wireguard{
		LinkAttrs: netlink.NewLinkAttrs(),
	}
	l.LinkAttrs.Name = "wg-test0"

	if err := netlink.LinkAdd(l); err != nil {
		t.Errorf("failed to create Wireguard interface: %s", err)
	}

	l2, err := netlink.LinkByName("wg-test0")
	if err != nil {
		t.Errorf("failed to get link details: %s", err)
	}

	if l2.Type() != "wireguard" {
		t.Fail()
	}

	if err := netlink.LinkDel(l); err != nil {
		t.Errorf("failed to delete Wireguard device: %w", err)
	}
}
