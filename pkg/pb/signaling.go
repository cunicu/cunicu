package pb

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"
	"riasc.eu/wice/pkg/crypto"
)

func (e *SignalingEnvelope) PublicKeyPair() (crypto.PublicKeyPair, error) {
	sender, err := crypto.ParseKeyBytes(e.Sender)
	if err != nil {
		return crypto.PublicKeyPair{}, fmt.Errorf("invalid key: %w", err)
	}

	recipient, err := crypto.ParseKeyBytes(e.Recipient)
	if err != nil {
		return crypto.PublicKeyPair{}, fmt.Errorf("invalid key: %w", err)
	}

	return crypto.PublicKeyPair{
		Ours:   recipient,
		Theirs: sender,
	}, nil
}

func (e *SignalingEnvelope) Decrypt(kp *crypto.KeyPair) (*SignalingMessage, error) {
	ekp, err := e.PublicKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys from envelope: %w", err)
	}

	if ekp != kp.Public() {
		return nil, errors.New("keypair mismatch")
	}

	msg := &SignalingMessage{}
	return msg, e.Contents.Unmarshal(msg, kp)
}

func (e *SignalingMessage) Encrypt(kp *crypto.KeyPair) (*SignalingEnvelope, error) {
	envp := &SignalingEnvelope{
		Sender:    kp.Ours.PublicKey().Bytes(),
		Recipient: kp.Theirs.Bytes(),
		Contents:  &EncryptedMessage{},
	}

	return envp, envp.Contents.Marshal(e, kp)
}

func (e *SignalingEnvelope) DeepCopyInto(out *SignalingEnvelope) {
	p := proto.Clone(e).(*SignalingEnvelope)
	*out = *p
}
