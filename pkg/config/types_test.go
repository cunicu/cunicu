// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"github.com/spf13/pflag"

	"cunicu.li/cunicu/pkg/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("types", func() {
	Context("output format", func() {
		for _, f := range config.OutputFormats {
			It(f.String(), func() {
				var g config.OutputFormat

				flags := pflag.NewFlagSet("test", pflag.ExitOnError)
				flags.Var(&g, "format", "Output format")

				err := flags.Parse([]string{"--format", f.String()})
				Expect(err).To(Succeed())
				Expect(g).To(BeEquivalentTo(f))

				h, err := f.MarshalText()
				Expect(err).To(Succeed())
				Expect(h).To(BeEquivalentTo(f.String()))
			})
		}

		It("fails on invalid format", func() {
			var g config.OutputFormat

			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			flags.Var(&g, "format", "Output format")

			err := flags.Parse([]string{"--format", "blub"})
			Expect(err).To(MatchError("invalid argument \"blub\" for \"--format\" flag: unknown output format: blub"))
		})
	})
})
