// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package os

import "kernel.org/pub/linux/libs/security/libcap/cap"

func HasCapabilities(caps ...cap.Value) bool {
	cs := cap.GetProc()

	for _, v := range caps {
		if s, err := cs.GetFlag(cap.Permitted, v); err != nil || !s {
			return false
		}
	}

	return true
}

func HasAdminPrivileges() bool {
	return HasCapabilities(cap.NET_ADMIN)
}
