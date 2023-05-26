// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build !linux

package os

func SetSysctl(_ string, _ any) error {
	return errNotSupported
}

func SetSysctlMap(_ map[string]any) error {
	return errNotSupported
}
