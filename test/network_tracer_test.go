//go:build tracer

package test_test

import (
	"bytes"

	"github.com/google/gopacket/pcapgo"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"riasc.eu/wice/pkg/wg/tracer"
)

type HandshakeTracer tracer.HandshakeTracer

func (n *Network) StartHandshakeTracer() {
	By("Starting WireGuard handshake tracer")

	tracer, err := tracer.NewHandshakeTracer()
	Expect(err).To(Succeed(), "Failed to setup WireGuard handshake tracer: %s", err)

	n.tracer = (*HandshakeTracer)(tracer)

	go func() {
		for {
			select {
			case hs := <-n.tracer.Handshakes:
				b := &bytes.Buffer{}
				err = hs.DumpKeyLog(b)
				Expect(err).To(Succeed(), "Failed to dump WireGuard handshake: %s", err)

				for _, c := range n.Network.Captures {
					err = c.WriteDecryptionSecret(pcapgo.DSB_SECRETS_TYPE_WIREGUARD, b.Bytes())
					Expect(err).To(Succeed(), "Failed to write decryption secrets to PCAPng file: %s", err)
				}

			case err := <-n.tracer.Errors:
				logger.Error("Failed to trace WireGuard handshake", zap.Error(err))
			}
		}
	}()
}

func (n *Network) StopHandshakeTracer() {
	if n.tracer != nil {
		By("Stopping WireGuard handshake tracer")

		err := (*tracer.HandshakeTracer)(n.tracer).Close()
		Expect(err).To(Succeed(), "Failed to close WireGuard handshake tracer; %s", err)
	}
}
