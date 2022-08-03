package test_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"riasc.eu/wice/pkg/test"
)

var _ = Describe("build", func() {
	It("finds base dir", func() {
		bd, err := test.FindBaseDir()
		Expect(err).To(Succeed())
		Expect(filepath.Join(bd, "etc/wice.yaml")).To(BeARegularFile())
	})

	It("can build wice binary", func() {
		if runtime.GOOS == "windows" {
			// TODO: Investigate
			Skip("This test is broken on Windows")
		}

		bin, err := test.BuildBinary(false)
		Expect(err).To(Succeed())

		fi, err := os.Stat(bin)
		Expect(err).To(Succeed())
		Expect(fi.Mode().IsRegular()).To(BeTrue())

		if runtime.GOOS != "windows" {
			Expect(fi.Mode() & 0100).NotTo(BeZero())
		}

		command := exec.Command(bin, "--help")
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).To(Succeed())
		Eventually(session.Out).Should(gbytes.Say(`WireGuard Interactive Connectivity Establishment`))
	})
})
