// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package os

import (
	"os"
	"os/signal"
	"syscall"
)

func SetupSignals(extraSignals ...os.Signal) chan os.Signal {
	signals := []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	signals = append(signals, extraSignals...)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)

	return ch
}
