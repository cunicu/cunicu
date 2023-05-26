// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build tracer

package e2e_test

import (
	"bytes"

	"github.com/gopacket/gopacket/pcapgo"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stv0g/cunicu/pkg/wg/tracer"
	"go.uber.org/zap"
)

type HandshakeTracer tracer.HandshakeTracer

func (n *Network) StartHandshakeTracer() {
	By("Starting WireGuard handshake tracer")

	tracer, err := tracer.NewHandshakeTracer()
	Expect(err).To(Succeed(), "Failed to setup WireGuard handshake tracer: %s", err)

	n.Tracer = (*HandshakeTracer)(tracer)

	go func() {
		for {
			select {
			case hs := <-n.Tracer.Handshakes:
				b := &bytes.Buffer{}
				err = hs.DumpKeyLog(b)
				Expect(err).To(Succeed(), "Failed to dump WireGuard handshake: %s", err)

				for _, c := range n.Network.Captures {
					err = c.WriteDecryptionSecret(pcapgo.DSB_SECRETS_TYPE_WIREGUARD, b.Bytes())
					Expect(err).To(Succeed(), "Failed to write decryption secrets to PCAPng file: %s", err)
				}

			case err := <-n.Tracer.Errors:
				logger.Error("Failed to trace WireGuard handshake", zap.Error(err))
			}
		}
	}()
}

func (n *Network) StopHandshakeTracer() {
	if n.Tracer != nil {
		By("Stopping WireGuard handshake tracer")

		err := (*tracer.HandshakeTracer)(n.Tracer).Close()
		Expect(err).To(Succeed(), "Failed to close WireGuard handshake tracer; %s", err)
	}
}
