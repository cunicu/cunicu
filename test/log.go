// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"

	"github.com/stv0g/cunicu/pkg/log"
	"github.com/stv0g/cunicu/pkg/tty"
)

func SetupLogging() *log.Logger {
	logger, err := SetupLoggingWithFile("", false)
	if err != nil {
		panic(err)
	}

	return logger
}

func SetupLoggingWithFile(fn string, truncate bool) (*log.Logger, error) {
	outputPaths := []string{"ginkgo"}

	if fn != "" {
		// Create parent directories for log file
		if path := filepath.Dir(fn); path != "" {
			if err := os.MkdirAll(path, 0o750); err != nil {
				panic(fmt.Errorf("failed to directory of log file: %w", err))
			}
		}

		fl := os.O_CREATE | os.O_APPEND | os.O_WRONLY
		if truncate {
			fl |= os.O_TRUNC
		}

		f, err := os.OpenFile(fn, fl, 0o644)
		if err != nil {
			panic(fmt.Errorf("failed to open log file '%s': %w", fn, err))
		}

		ginkgo.GinkgoWriter.TeeTo(tty.NewANSIStripper(f))
	}

	return log.SetupLogging("", outputPaths, true)
}
