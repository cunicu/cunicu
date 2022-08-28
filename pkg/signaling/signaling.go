// Package signaling implements various signaling backends to exchange encrypted messages between peers
package signaling

import (
	"riasc.eu/wice/pkg/crypto"

	signalingproto "riasc.eu/wice/pkg/proto/signaling"
)

type Message = signalingproto.Message
type Envelope = signalingproto.Envelope

type MessageHandler interface {
	OnSignalingMessage(*crypto.PublicKeyPair, *Message)
}

type EnvelopeHandler interface {
	OnSignalingEnvelope(*Envelope)
}
