package signaling

import (
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
)

type Message = pb.SignalingMessage
type Envelope = pb.SignalingEnvelope

type MessageHandler interface {
	OnSignalingMessage(*crypto.PublicKeyPair, *Message)
}

type EnvelopeHandler interface {
	OnSignalingEnvelope(*Envelope)
}
