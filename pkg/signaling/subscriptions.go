package signaling

import (
	"errors"
	"fmt"
	"sync"

	"go.uber.org/zap"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
)

type Subscription struct {
	kp      *crypto.KeyPair
	Channel chan *pb.SignalingMessage
}

type SubscriptionsRegistry struct {
	subs     map[crypto.Key]*Subscription
	subsLock sync.RWMutex
}

func NewSubscriptionsRegistry() SubscriptionsRegistry {
	return SubscriptionsRegistry{
		subs: map[crypto.Key]*Subscription{},
	}
}

func (s *SubscriptionsRegistry) NewMessage(env *pb.SignalingEnvelope) error {
	sender, err := crypto.ParseKeyBytes(env.Sender)
	if err != nil {
		return fmt.Errorf("invalid key: %w", err)
	}

	sub, err := s.GetSubscription(&sender)
	if err != nil {
		return err
	}

	return sub.NewMessage(env)
}

func (s *SubscriptionsRegistry) NewSubscription(kp *crypto.KeyPair) (*Subscription, error) {
	s.subsLock.Lock()
	defer s.subsLock.Unlock()

	if _, ok := s.subs[kp.Theirs]; ok {
		return nil, errors.New("already existing")
	}

	sub := &Subscription{
		kp:      kp,
		Channel: make(chan *pb.SignalingMessage, 100),
	}

	s.subs[kp.Theirs] = sub

	return sub, nil
}

func (s *SubscriptionsRegistry) GetSubscription(pk *crypto.Key) (*Subscription, error) {
	s.subsLock.Lock()
	defer s.subsLock.Unlock()

	sub, ok := s.subs[*pk]
	if !ok {
		return nil, errors.New("missing subscription")
	}

	return sub, nil
}

func (s *SubscriptionsRegistry) Unsubscribe(kp *crypto.KeyPair) {
	s.subsLock.Lock()
	defer s.subsLock.Unlock()

	sub, ok := s.subs[kp.Theirs]
	if !ok {
		return
	}

	sub.Close()

	delete(s.subs, kp.Theirs)
}

func (s *Subscription) Close() error {
	close(s.Channel)

	return nil
}

func (s *Subscription) NewMessage(env *pb.SignalingEnvelope) error {
	msg, err := env.Decrypt(s.kp)
	if err != nil {
		return err
	}

	zap.L().Named("backend").Debug("Received signaling message", zap.Any("msg", msg), zap.Any("kp", s.kp))

	s.Channel <- msg

	return nil
}
