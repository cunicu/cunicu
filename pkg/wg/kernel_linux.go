// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package wg

import (
	"errors"
	"math"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

func KernelModuleExists() bool {
	l := &netlink.Wireguard{
		LinkAttrs: netlink.NewLinkAttrs(),
	}
	l.LinkAttrs.Name = "must-not-exist"

	// We willingly try to create a device with an invalid
	// MTU here as the validation of the MTU will be performed after
	// the validation of the link kind and hence allows us to check
	// for the existence of the WireGuard module without actually
	// creating a link.
	//
	// As a side-effect, this will also let the kernel lazy-load
	// the WireGuard module.
	l.LinkAttrs.MTU = math.MaxInt

	err := netlink.LinkAdd(l)

	return errors.Is(err, unix.EINVAL)
}
