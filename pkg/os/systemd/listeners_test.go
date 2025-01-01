// SPDX-FileCopyrightText: 2015 CoreOS, Inc.
// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build linux

package systemd_test

import (
	"net"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Context("Listeners", func() {
	var cmd *exec.Cmd

	BeforeEach(func() {
		path, err := Build("../../../test/systemd/listen.go")
		Expect(err).To(Succeed())

		cmd = exec.Command(path)
	})

	// Forks out a copy of activation.go example and reads back two
	// strings from the pipes that are passed in.
	It("can pass listeners", func() {
		t1, err := net.ListenTCP("tcp", &net.TCPAddr{Port: 9999})
		Expect(err).To(Succeed())

		t2, err := net.ListenTCP("tcp", &net.TCPAddr{Port: 1234})
		Expect(err).To(Succeed())

		f1, err := t1.File()
		Expect(err).To(Succeed())

		f2, err := t2.File()
		Expect(err).To(Succeed())

		cmd.ExtraFiles = []*os.File{f1, f2}

		r1, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: net.IPv6loopback, Port: 9999})
		Expect(err).To(Succeed())

		_, err = r1.Write([]byte("Hi"))
		Expect(err).To(Succeed())

		r2, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: net.IPv6loopback, Port: 1234})
		Expect(err).To(Succeed())

		_, err = r2.Write([]byte("Hi"))
		Expect(err).To(Succeed())

		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "LISTEN_FDS=2", "LISTEN_FDNAMES=fd1:fd2", "FIX_LISTEN_PID=1")

		session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).To(Succeed())
		Eventually(session).Should(Exit(0))

		Eventually(BufferReader(r1)).Should(Say("Hello world: fd1"))
		Eventually(BufferReader(r2)).Should(Say("Goodbye world: fd2"))
	})
})
