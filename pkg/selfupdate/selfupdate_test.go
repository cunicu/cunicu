// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package selfupdate_test

import (
	"path/filepath"
	"testing"

	"github.com/stv0g/cunicu/pkg/buildinfo"
	"github.com/stv0g/cunicu/pkg/log"
	"github.com/stv0g/cunicu/pkg/selfupdate"
	"github.com/stv0g/cunicu/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	test.SetupLogging()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Self-update Suite")
}

var _ = It("self-update", Pending, func() {
	logger := log.Global.Named("self-update")

	output := filepath.Join(GinkgoT().TempDir(), "cunicu")

	// We fake a lower version for force an update
	buildinfo.Version = "0.0.9"

	_, err := selfupdate.SelfUpdate(output, logger)
	Expect(err).To(Succeed())
})
