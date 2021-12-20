//go:build linux
// +build linux

package main_test

import (
	"net"

	"github.com/stv0g/gont"
)

func main() {
	n := gont.NewNetwork("test")

	n.Reset()

	sw1, _ := n.AddSwitch("sw1")
	sw2, _ := n.AddSwitch("sw2")
	sw3, _ := n.AddSwitch("sw3")

	mask := net.IPv4Mask(255, 255, 255, 0)

	h12, _ := n.AddHost("h12", net.IPv4(10, 0, 1, 1),
		&gont.Interface{"eth0", net.IPv4(10, 0, 1, 2), mask, sw1})

	h22, _ := n.AddHost("h22", net.IPv4(10, 0, 2, 1),
		&gont.Interface{"eth0", net.IPv4(10, 0, 2, 2), mask, sw2})

	n.AddHost("h23", net.IPv4(10, 0, 2, 1),
		&gont.Interface{"eth0", net.IPv4(10, 0, 2, 3), mask, sw2})

	h32, _ := n.AddHost("h32", net.IPv4(10, 0, 3, 1),
		&gont.Interface{"eth0", net.IPv4(10, 0, 3, 2), mask, sw3})

	n.AddNAT("n1", nil,
		&gont.Interface{"nb", net.IPv4(10, 0, 1, 1), mask, sw1},
		&gont.Interface{"sb", net.IPv4(10, 0, 2, 1), mask, sw2})

	n.AddNAT("n2", nil,
		&gont.Interface{"nb", net.IPv4(10, 0, 2, 10), mask, sw2},
		&gont.Interface{"sb", net.IPv4(10, 0, 3, 1), mask, sw3})

	h32.Ping(h12)
	h22.Traceroute(h12)
}
