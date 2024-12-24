// SPDX-FileCopyrightText: 2016 CoreOS, Inc.
// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build linux

package systemd_test

import (
	"net"
	"os"
	"path/filepath"

	"cunicu.li/cunicu/pkg/os/systemd"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("Notify", func() {
	var testDir, notifySocket string
	var conn *net.UnixConn
	var err error

	BeforeEach(func() {
		testDir = GinkgoT().TempDir()

		notifySocket = filepath.Join(testDir, "notify-socket.sock")

		conn, err = net.ListenUnixgram("unixgram", &net.UnixAddr{
			Name: notifySocket,
			Net:  "unixgram",
		})
		Expect(err).To(Succeed())
	})

	AfterEach(func() {
		err = conn.Close()
		Expect(err).To(Succeed())
	})

	DescribeTable("works", func(unsetEnv bool, envCb func() string, exptectSent bool, expectErr error) {
		env := envCb()
		err = os.Setenv("NOTIFY_SOCKET", env)
		Expect(err).To(Succeed())

		sent, err := systemd.Notify(unsetEnv, systemd.NotifyReady)

		if expectErr != nil {
			Expect(err).To(MatchError(err))
		} else {
			Expect(err).To(Succeed())
		}

		Expect(sent).To(Equal(exptectSent))

		if unsetEnv && env != "" {
			Expect(os.Getenv("NOTIFY_SOCKET")).To(BeEmpty())
		}
	},
		Entry("Notification supported, data has been sent: (true, nil)", false, func() string { return notifySocket }, true, nil),
		Entry("Notification supported, but failure happened: (false, err)", true, func() string { return filepath.Join(testDir, "missing.sock") }, false, os.ErrClosed),
		Entry("Notification not supported: (false, nil)", true, func() string { return "" }, false, nil),
	)
})
