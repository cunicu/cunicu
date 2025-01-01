// SPDX-FileCopyrightText: 2015 CoreOS, Inc.
// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build linux

package systemd_test

import (
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Context("Files", func() {
	var cmd *exec.Cmd

	BeforeEach(func() {
		path, err := Build("../../../test/systemd/activation.go")
		Expect(err).To(Succeed())

		cmd = exec.Command(path)
	})

	// Forks out a copy of activation.go example and reads back two
	// strings from the pipes that are passed in.
	It("can pass files as FDs", func() {
		r1, w1, _ := os.Pipe()
		r2, w2, _ := os.Pipe()
		cmd.ExtraFiles = []*os.File{w1, w2}

		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "LISTEN_FDS=2", "LISTEN_FDNAMES=fd1", "FIX_LISTEN_PID=1")

		session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).To(Succeed())
		Eventually(session).Should(Exit(0))

		Eventually(BufferReader(r1)).Should(Say("Hello world: fd1"))
		Eventually(BufferReader(r2)).Should(Say("Goodbye world: LISTEN_FD_4"))
	})

	It("fails when FIX_LISTEN_PID is not set", func() {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "LISTEN_FDS=2")

		session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).To(Succeed())

		Eventually(session).Should(Exit(2))

		Expect(session.Err).To(Say("No files"))
	})

	It("fails when no FDs are passed ", func() {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "LISTEN_FDS=0", "FIX_LISTEN_PID=1")

		session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).To(Succeed())
		Eventually(session).Should(Exit(2))

		Expect(session.Err).To(Say("No files"))
	})
})
