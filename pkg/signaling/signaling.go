// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package signaling implements various signaling backends to exchange encrypted messages between peers
package signaling

import (
	"github.com/stv0g/cunicu/pkg/crypto"
	signalingproto "github.com/stv0g/cunicu/pkg/proto/signaling"
)

type (
	Message  = signalingproto.Message
	Envelope = signalingproto.Envelope
)

type MessageHandler interface {
	OnSignalingMessage(*crypto.PublicKeyPair, *Message)
}

type EnvelopeHandler interface {
	OnSignalingEnvelope(*Envelope)
}
