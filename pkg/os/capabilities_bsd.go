// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build darwin || freebsd || openbsd || netbsd || plan9 || dragonfly || solaris

package os

import (
	"os"
)

func HasAdminPrivileges() bool {
	return os.Geteuid() == 0
}
