// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build test

package main

import (
	"testing"
)

func TestRunMain(t *testing.T) {
	wgCmd.Execute()
}
