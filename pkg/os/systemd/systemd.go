// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package systemd provides a Go implementation of systemd related protocols.
//
// Currently, the followwing features are supported:
//
//   - sd_notify: It can be used to inform systemd of service start-up completion,
//     watchdog events, and other status changes.
//     See: https://www.freedesktop.org/software/systemd/man/sd_notify.html#Description
package systemd
