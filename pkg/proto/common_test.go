// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto_test

import (
	"time"

	"github.com/stv0g/cunicu/pkg/proto"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("common", func() {
	It("time", func() {
		now := time.Now()
		protonow := proto.Time(now)

		Expect(protonow.Time()).To(BeTemporally("==", now))
	})
})
