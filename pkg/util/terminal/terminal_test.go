package terminal_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stv0g/cunicu/pkg/util/terminal"
	"github.com/stv0g/cunicu/test"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Terminal Suite")
}

var _ = test.SetupLogging()

var _ = Context("tty", func() {
	It("is true", func() {
		Expect(terminal.IsATTY(os.Stdout)).To(BeTrue())
	})

	It("is false", func() {
		fn := filepath.Join(GinkgoT().TempDir(), "file")
		f, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY, 0600)
		Expect(err).To(Succeed())

		Expect(terminal.IsATTY(f)).To(BeFalse())
	})
})
