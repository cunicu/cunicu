// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package selfupdate_test

import (
	"context"
	"time"

	"github.com/stv0g/cunicu/pkg/selfupdate"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("github", func() {
	It("can get latest release", func() {
		rel, err := selfupdate.GitHubLatestRelease(context.Background())
		Expect(err).To(Succeed())

		Expect(rel.Version).To(MatchRegexp(`\d+\.\d+\.\d+`))
		Expect(rel.PublishedAt).To(BeTemporally("<", time.Now()))
		Expect(len(rel.Assets)).To(BeNumerically(">", 10))
	})
})
