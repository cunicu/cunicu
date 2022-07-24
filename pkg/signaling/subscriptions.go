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
	onMessages    map[crypto.Key][]MessageHandler
	onAllMessages []MessageHandler

	sk crypto.Key
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
	pk, err := crypto.ParseKeyBytes(env.Recipient)
	if err != nil {
		return fmt.Errorf("invalid key: %w", err)
	}

	sub, err := s.GetSubscription(&pk)
	if err != nil {
		return err
	}

	return sub.NewMessage(env)
}

func (s *SubscriptionsRegistry) NewSubscription(k *crypto.Key) (*Subscription, error) {
	s.subsLock.Lock()
	defer s.subsLock.Unlock()

	if _, ok := s.subs[k.PublicKey()]; ok {
		return nil, errors.New("already existing")
	}

	sub := &Subscription{
		onMessages:    map[crypto.Key][]MessageHandler{},
		onAllMessages: []MessageHandler{},
		sk:            *k,
	}

	s.subs[k.PublicKey()] = sub

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

func (s *SubscriptionsRegistry) GetOrCreateSubscription(sk *crypto.Key) (bool, *Subscription, error) {
	pk := sk.PublicKey()

	sub, err := s.GetSubscription(&pk)
	if err == nil {
		return false, sub, nil
	}

	sub, err = s.NewSubscription(sk)

	return true, sub, err
}

func (s *SubscriptionsRegistry) SubscribeAll(sk *crypto.Key, h MessageHandler) (bool, error) {
	created, sub, err := s.GetOrCreateSubscription(sk)
	if err != nil {
		return false, err
	}

	sub.onAllMessages = append(sub.onAllMessages, h)

	return created, nil
}

func (s *SubscriptionsRegistry) Subscribe(kp *crypto.KeyPair, h MessageHandler) (bool, error) {
	created, sub, err := s.GetOrCreateSubscription(&kp.Ours)
	if err != nil {
		return false, err
	}

	sub.OnMessages(&kp.Theirs, h)

	return created, nil
}

func (s *SubscriptionsRegistry) Unsubscribe(pk *crypto.Key) {
	s.subsLock.Lock()
	defer s.subsLock.Unlock()

	delete(s.subs, *pk)
}

func (s *Subscription) NewMessage(env *pb.SignalingEnvelope) error {
	pk, err := crypto.ParseKeyBytes(env.Sender)
	if err != nil {
		return fmt.Errorf("failed to parse sender key: %w", err)
	}

	kp := crypto.KeyPair{
		Ours:   s.sk,
		Theirs: pk,
	}
	pkp := kp.Public()

	msg, err := env.Decrypt(&kp)
	if err != nil {
		return err
	}

	zap.L().Named("backend").Debug("Received signaling message", zap.Any("msg", msg), zap.Any("pkp", pkp))

	for _, cb := range s.onAllMessages {
		cb.OnSignalingMessage(&pkp, msg)
	}

	if cbs, ok := s.onMessages[kp.Theirs]; ok {
		for _, cb := range cbs {
			cb.OnSignalingMessage(&pkp, msg)
		}
	}

	return nil
}

func (s *Subscription) OnMessages(pk *crypto.Key, h MessageHandler) {
	if _, ok := s.onMessages[*pk]; ok {
		s.onMessages[*pk] = append(s.onMessages[*pk], h)
	} else {
		s.onMessages[*pk] = []MessageHandler{h}
	}
}

func (s *Subscription) OnAllMessages(h MessageHandler) {
	s.onAllMessages = append(s.onAllMessages, h)
}
