// SPDX-FileCopyrightText: 2015 CoreOS, Inc.
// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build !linux

package systemd

import (
	"os"
)

// Files returns a slice containing a `os.File` object for each
// file descriptor passed to this process via systemd fd-passing protocol.
//
// The order of the file descriptors is preserved in the returned slice.
// `unsetEnv` is typically set to `true` in order to avoid clashes in
// fd usage and to avoid leaking environment flags to child processes.
func Files(_ bool) []*os.File {
	return nil
}

func NumFiles() int {
	return 0
}
