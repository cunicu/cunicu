package main_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	g "github.com/stv0g/gont/pkg"
	"github.com/vishvananda/netlink"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/rpc"
	"riasc.eu/wice/pkg/test"
	"riasc.eu/wice/pkg/util"
)

var _ = Describe("single isolated host", func() {
	var err error

	var n *g.Network
	var h1 *g.Host

	BeforeEach(func() {
		if !util.HasAdminPrivileges() {
			Skip("Insufficient privileges")
		}

		n, err = g.NewNetwork("")
		Expect(err).To(Succeed())

		h1, err = n.AddHost("h1")
		Expect(err).To(Succeed())
	})

	AfterEach(func() {
		fmt.Println("Exited")

		Expect(n.Close()).To(Succeed())
	})

	Context("start without any existing interfaces", func() {
		It("has only a loopback interface", func() {
			h1.RunFunc(func() error {
				links, err := netlink.LinkList()
				Expect(err).To(Succeed())
				Expect(links).To(HaveLen(1))
				Expect(links[0].Attrs().Name).To(Equal("lo"))

				return nil
			})
		})

		Context("wice", func() {
			var cmd *exec.Cmd
			var client *rpc.Client
			var tmpDir string

			BeforeEach(func() {
				tmpDir = GinkgoT().TempDir()
				sockPath := filepath.Join(tmpDir, "wice.sock")

				_, _, cmd, err = test.StartWiceWithCoverage(h1, "daemon",
					"--log-level", "debug",
					"--socket", sockPath)
				Expect(err).To(Succeed())

				Eventually(func() error {
					client, err = rpc.Connect(sockPath)
					return err
				}).Should(Succeed(), "failed to connect to control socket: %w", err)
			})

			AfterEach(func() {
				err = cmd.Process.Signal(os.Interrupt)
				Expect(err).To(Succeed())

				err = client.Close()
				Expect(err).To(Succeed())

				err = os.RemoveAll(tmpDir)
				Expect(err).To(Succeed())
			})

			It("returns no interfaces in status", func() {
				h1.RunFunc(func() error {
					sts, err := client.GetStatus(context.Background(), &pb.Void{})
					Expect(err).To(Succeed())
					Expect(sts.Interfaces).To(BeEmpty())

					return nil
				})
			})
		})
	})
})
