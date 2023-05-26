// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build !linux

package config

const (
	// FreeBSD FIBs start at 0
	// See: https://www.freebsd.org/cgi/man.cgi?query=setfib&sektion=2&n=1
	DefaultRouteTable = 0
)
