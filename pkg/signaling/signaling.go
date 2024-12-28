// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package signaling implements various signaling backends to exchange encrypted messages between peers
package signaling

import (
	"cunicu.li/cunicu/pkg/crypto"
	signalingproto "cunicu.li/cunicu/pkg/proto/signaling"
)

type (
	Message  = signalingproto.Message
	Envelope = signalingproto.Envelope
)

type MessageHandler interface {
	OnSignalingMessage(kp *crypto.PublicKeyPair, msg *Message)
}

type EnvelopeHandler interface {
	OnSignalingEnvelope(env *Envelope)
}
