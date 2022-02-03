package pb

import (
	"errors"

	"riasc.eu/wice/pkg/crypto"
)

func (m *SignalingEnvelope) Decrypt(kp *crypto.KeyPair) (*SignalingMessage, error) {
	var sender = *(*crypto.Key)(m.Sender)
	var receipient = *(*crypto.Key)(m.Receipient)

	if receipient != kp.Ours.PublicKey() {
		return nil, errors.New("invalid receipient key")
	}

	if sender != kp.Theirs {
		return nil, errors.New("invalid sender key")
	}

	msg := &SignalingMessage{}
	return msg, m.Contents.Unmarshal(msg, kp)
}

func (m *SignalingMessage) Encrypt(kp *crypto.KeyPair) (*SignalingEnvelope, error) {
	envp := &SignalingEnvelope{
		Sender:     kp.Ours.PublicKey().Bytes(),
		Receipient: kp.Theirs.Bytes(),
		Contents:   &EncryptedMessage{},
	}

	return envp, envp.Contents.Marshal(m, kp)
}
