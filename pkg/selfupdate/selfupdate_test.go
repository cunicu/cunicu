// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package selfupdate_test

import (
	"path/filepath"
	"testing"

	"cunicu.li/cunicu/pkg/buildinfo"
	"cunicu.li/cunicu/pkg/log"
	"cunicu.li/cunicu/pkg/selfupdate"
	"cunicu.li/cunicu/test"

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
