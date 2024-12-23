// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build unix

package config

const DefaultSocketPath = "/run/cunicu.sock"

//nolint:gochecknoglobals
var RuntimeConfigFile = "/var/lib/cunicu/runtime.yaml"
