// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build windows

package os

import (
	"syscall"
)

const (
	SigUpdate = syscall.Signal(-1) // not supported
	SigReload = syscall.Signal(-2) // not supported
)
