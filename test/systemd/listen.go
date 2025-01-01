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

	listenersWithNames, err := systemd.ListenersWithNames()
	if err != nil {
		panic(err)
	}

	if os.Getenv("LISTEN_PID") != "" || os.Getenv("LISTEN_FDS") != "" || os.Getenv("LISTEN_FDNAMES") != "" {
		panic("Can not unset envs")
	}

	c0, _ := listenersWithNames["fd1"][0].Accept()
	c1, _ := listenersWithNames["fd2"][0].Accept()

	// Write out the expected strings to the two pipes
	c0.Write([]byte("Hello world: fd1"))
	c1.Write([]byte("Goodbye world: fd2"))

	return
}
