// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build !tracer

package e2e_test

type HandshakeTracer any

func (n *Network) StartHandshakeTracer() {}
func (n *Network) StopHandshakeTracer()  {}
