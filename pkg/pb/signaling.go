package pb

import (
	"errors"
	"fmt"

	"riasc.eu/wice/pkg/crypto"
)

func (m *SignalingEnvelope) Decrypt(kp *crypto.KeyPair) (*SignalingMessage, error) {
	sender, err := crypto.ParseKeyBytes(m.Sender)
	if err != nil {
		return nil, fmt.Errorf("invalid key: %w", err)
	}
	receipient, err := crypto.ParseKeyBytes(m.Receipient)
	if err != nil {
		return nil, fmt.Errorf("invalid key: %w", err)
	}

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
