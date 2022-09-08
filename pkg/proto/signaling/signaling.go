package signaling

import (
	"errors"
	"fmt"

	"github.com/stv0g/cunicu/pkg/crypto"
	"golang.org/x/crypto/nacl/box"
	"google.golang.org/protobuf/proto"
)

func (e *Envelope) PublicKeyPair() (crypto.PublicKeyPair, error) {
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

func (e *Envelope) Decrypt(kp *crypto.KeyPair) (*Message, error) {
	ekp, err := e.PublicKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys from envelope: %w", err)
	}

	if ekp != kp.Public() {
		return nil, errors.New("key pair mismatch")
	}

	msg := &Message{}
	return msg, e.Contents.Unmarshal(msg, kp)
}

func (e *Message) Encrypt(kp *crypto.KeyPair) (*Envelope, error) {
	envp := &Envelope{
		Sender:    kp.Ours.PublicKey().Bytes(),
		Recipient: kp.Theirs.Bytes(),
		Contents:  &EncryptedMessage{},
	}

	return envp, envp.Contents.Marshal(e, kp)
}

func (e *Envelope) DeepCopyInto(out *Envelope) {
	p := proto.Clone(e).(*Envelope)
	*out = *p
}

func (s *EncryptedMessage) Marshal(msg proto.Message, kp *crypto.KeyPair) error {
	body, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	s.Nonce, err = crypto.GetNonce(24)
	if err != nil {
		return fmt.Errorf("failed to create nonce: %w", err)
	}

	s.Body = box.Seal([]byte{}, body, (*[24]byte)(s.Nonce), (*[32]byte)(&kp.Theirs), (*[32]byte)(&kp.Ours))
	if err != nil {
		return fmt.Errorf("failed to seal: %w", err)
	}

	return nil
}

func (s *EncryptedMessage) Unmarshal(msg proto.Message, kp *crypto.KeyPair) error {
	if len(s.Nonce) != 24 {
		return errors.New("invalid nonce length")
	}

	body, ok := box.Open([]byte{}, s.Body, (*[24]byte)(s.Nonce), (*[32]byte)(&kp.Theirs), (*[32]byte)(&kp.Ours))
	if !ok {
		return errors.New("failed to open")
	}

	return proto.Unmarshal(body, msg)
}
