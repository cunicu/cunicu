package link_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	osx "github.com/stv0g/cunicu/pkg/os"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Link Suite")
}

var _ = BeforeSuite(func() {
	if !osx.HasAdminPrivileges() {
		Skip("Insufficient privileges")
	}
})
