package pb

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/nacl/box"
	"google.golang.org/protobuf/proto"
	"riasc.eu/wice/pkg/crypto"
)

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
