// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build !(linux || darwin || freebsd || openbsd || netbsd || plan9 || dragonfly || solaris)

package os

func HasAdminPrivileges() bool {
	return false
}
