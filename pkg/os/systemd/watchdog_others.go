// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build !linux

package systemd

import "time"

func WatchdogEnabled(_ bool) (interval time.Duration, err error) {
	return 0, nil
}
