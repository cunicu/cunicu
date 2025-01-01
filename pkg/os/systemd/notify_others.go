// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build !linux

package systemd

func Notify(_ bool, _ ...string) (bool, error) {
	return false, nil
}
