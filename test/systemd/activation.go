// SPDX-FileCopyrightText: 2015 CoreOS, Inc.
// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build ignore

// Activation example used by the activation unit tests.
package main

import (
	"fmt"
	"os"

	"cunicu.li/cunicu/pkg/os/systemd"
)

func fixListenPid() {
	if os.Getenv("FIX_LISTEN_PID") != "" {
		// HACK: real systemd would set LISTEN_PID before exec'ing but
		// this is too difficult in golang for the purpose of a test.
		// Do not do this in real code.
		os.Setenv("LISTEN_PID", fmt.Sprintf("%d", os.Getpid()))
	}
}

func main() {
	fixListenPid()

	files := systemd.Files(false)

	if len(files) == 0 {
		panic("No files")
	}

	if os.Getenv("LISTEN_PID") == "" || os.Getenv("LISTEN_FDS") == "" || os.Getenv("LISTEN_FDNAMES") == "" {
		panic("Should not unset envs")
	}

	files = systemd.Files(true)

	if os.Getenv("LISTEN_PID") != "" || os.Getenv("LISTEN_FDS") != "" || os.Getenv("LISTEN_FDNAMES") != "" {
		panic("Can not unset envs")
	}

	// Write out the expected strings to the two pipes
	files[0].Write([]byte("Hello world: " + files[0].Name()))
	files[1].Write([]byte("Goodbye world: " + files[1].Name()))

	return
}
