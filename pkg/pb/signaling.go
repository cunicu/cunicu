package pb

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"
	"riasc.eu/wice/pkg/crypto"
)

func (e *SignalingEnvelope) Decrypt(kp *crypto.KeyPair) (*SignalingMessage, error) {
	sender, err := crypto.ParseKeyBytes(e.Sender)
	if err != nil {
		return nil, fmt.Errorf("invalid key: %w", err)
	}
	receipient, err := crypto.ParseKeyBytes(e.Receipient)
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
	return msg, e.Contents.Unmarshal(msg, kp)
}

func (e *SignalingMessage) Encrypt(kp *crypto.KeyPair) (*SignalingEnvelope, error) {
	envp := &SignalingEnvelope{
		Sender:     kp.Ours.PublicKey().Bytes(),
		Receipient: kp.Theirs.Bytes(),
		Contents:   &EncryptedMessage{},
	}

	return envp, envp.Contents.Marshal(e, kp)
}

func (e *SignalingEnvelope) DeepCopyInto(out *SignalingEnvelope) {
	p := proto.Clone(e).(*SignalingEnvelope)
	*out = *p
}
