// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build unix

package os

import (
	"os"

	"golang.org/x/sys/unix"
)

const ReexecSelfSupported = true

func ReexecSelf() error {
	return unix.Exec("/proc/self/exe", os.Args, os.Environ())
}
