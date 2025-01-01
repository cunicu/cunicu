// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build !windows

package os

import (
	"golang.org/x/sys/unix"
)

const (
	SigUpdate = unix.SIGUSR1
	SigReload = unix.SIGHUP
)
