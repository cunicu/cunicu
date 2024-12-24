// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build !unix

package config

const DefaultSocketPath = "cunicu.sock"

var RuntimeConfigFile = "runtime.yaml"
