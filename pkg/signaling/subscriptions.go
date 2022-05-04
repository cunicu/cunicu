package signaling

import (
	"errors"
	"fmt"
	"sync"

	"go.uber.org/zap"
	"riasc.eu/wice/internal/types"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
)

type Subscription struct {
	*types.Fanout[*pb.SignalingMessage]

	kp crypto.KeyPair
}

type SubscriptionsRegistry struct {
	subs     map[crypto.PublicKeyPair]*Subscription
	subsLock sync.RWMutex
}

func NewSubscriptionsRegistry() SubscriptionsRegistry {
	return SubscriptionsRegistry{
		subs: map[crypto.PublicKeyPair]*Subscription{},
	}
}

func (s *SubscriptionsRegistry) NewMessage(env *pb.SignalingEnvelope) error {
	sender, err := crypto.ParseKeyBytes(env.Sender)
	if err != nil {
		return fmt.Errorf("invalid key: %w", err)
	}

	receipient, err := crypto.ParseKeyBytes(env.Receipient)
	if err != nil {
		return fmt.Errorf("invalid key: %w", err)
	}

	pkp := crypto.PublicKeyPair{
		Ours:   receipient,
		Theirs: sender,
	}

	sub, err := s.GetSubscription(&pkp)
	if err != nil {
		return err
	}

	return sub.NewMessage(env)
}

func (s *SubscriptionsRegistry) NewSubscription(kp *crypto.KeyPair) (*Subscription, error) {
	s.subsLock.Lock()
	defer s.subsLock.Unlock()

	if _, ok := s.subs[kp.Public()]; ok {
		return nil, errors.New("already existing")
	}

	sub := &Subscription{
		Fanout: types.NewFanout[*pb.SignalingMessage](),
		kp:     *kp,
	}

	s.subs[kp.Public()] = sub

	return sub, nil
}

func (s *SubscriptionsRegistry) GetSubscription(pkp *crypto.PublicKeyPair) (*Subscription, error) {
	s.subsLock.Lock()
	defer s.subsLock.Unlock()

	sub, ok := s.subs[*pkp]
	if !ok {
		return nil, errors.New("missing subscription")
	}

	return sub, nil
}

func (s *SubscriptionsRegistry) GetSubscriptions() ([]crypto.PublicKeyPair, error) {
	s.subsLock.Lock()
	defer s.subsLock.Unlock()

	pkps := []crypto.PublicKeyPair{}
	for pkp := range s.subs {
		pkps = append(pkps, pkp)
	}

	return pkps, nil
}

func (s *SubscriptionsRegistry) GetOrCreateSubscription(kp *crypto.KeyPair) (*Subscription, error) {
	pkp := kp.Public()

	sub, err := s.GetSubscription(&pkp)
	if err == nil {
		return sub, nil
	}

	return s.NewSubscription(kp)
}

func (s *SubscriptionsRegistry) Subscribe(kp *crypto.KeyPair) (chan *pb.SignalingMessage, error) {
	sub, err := s.GetOrCreateSubscription(kp)
	if err != nil {
		return nil, err
	}

	return sub.AddChannel(), nil
}

func (s *SubscriptionsRegistry) Unsubscribe(pkp *crypto.PublicKeyPair) {
	s.subsLock.Lock()
	defer s.subsLock.Unlock()

	sub, ok := s.subs[*pkp]
	if !ok {
		return
	}

	sub.Close()

	delete(s.subs, *pkp)
}

func (s *Subscription) NewMessage(env *pb.SignalingEnvelope) error {
	msg, err := env.Decrypt(&s.kp)
	if err != nil {
		return err
	}

	zap.L().Named("backend").Debug("Received signaling message", zap.Any("msg", msg), zap.Any("kp", s.kp))

	s.C <- msg

	return nil
}
