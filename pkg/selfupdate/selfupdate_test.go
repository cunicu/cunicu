// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package selfupdate_test

import (
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stv0g/cunicu/pkg/buildinfo"
	"github.com/stv0g/cunicu/pkg/selfupdate"
	"github.com/stv0g/cunicu/test"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Self-update Suite")
}

var logger = test.SetupLogging()

var _ = It("self-update", Pending, func() {
	logger := logger.Named("self-update")

	output := filepath.Join(GinkgoT().TempDir(), "cunicu")

	// We fake a lower version for force an update
	buildinfo.Version = "0.0.9"

	_, err := selfupdate.SelfUpdate(output, logger)
	Expect(err).To(Succeed())
})
