// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build windows

package os

import (
	"syscall"
)

const (
	SigUpdate = syscall.Signal(-1) // not supported
)
