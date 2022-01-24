package pb

import (
	"crypto/rand"
	"errors"
	"fmt"

	"golang.org/x/crypto/nacl/box"
	"google.golang.org/protobuf/proto"
	"riasc.eu/wice/pkg/crypto"
)

func (m *SignedMessage) Validate(pk crypto.Key) (bool, error) {
	if len(m.Signature) != crypto.SignatureLength {
		return false, errors.New("invalid signature length")
	}

	sig := *(*crypto.Signature)(m.Signature)

	return pk.Verify(m.Body, sig), nil
}

func (s *SignedMessage) Marshal(msg proto.Message, sk crypto.Key) error {
	var err error

	s.Body, err = proto.Marshal(msg)
	if err != nil {
		return err
	}

	nonce, err := crypto.GetNonce()
	if err != nil {
		return fmt.Errorf("failed to generate nonce: %s", err)
	}

	sig := sk.Sign(s.Body, nonce)

	s.Signature = sig[:]

	return nil
}

func (s *SignedMessage) Unmarshal(msg proto.Message, pk crypto.Key) error {
	if ok, err := s.Validate(pk); err != nil {
		return err
	} else if !ok {
		return errors.New("invalid signature")
	} else {
		return proto.Unmarshal(s.Body, msg)
	}
}

func (s *EncryptedMessage) Marshal(msg proto.Message, pk crypto.Key) error {
	body, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	s.Body, err = box.SealAnonymous([]byte{}, body, (*[32]byte)(&pk), rand.Reader)
	if err != nil {
		fmt.Errorf("failed to seal: %w", err)
	}

	return nil
}

func (s *EncryptedMessage) Unmarshal(msg proto.Message, sk crypto.Key) error {
	body := []byte{}

	pk := sk.PublicKey()

	_, ok := box.OpenAnonymous(body, s.Body, (*[32]byte)(&pk), (*[32]byte)(&sk))
	if !ok {
		return errors.New("failed to open")
	}

	return proto.Unmarshal(body, msg)
}
