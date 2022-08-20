package pske

import (
	"context"

	kyber "github.com/symbolicsoft/kyber-k2so"
	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/signaling"

	pskeproto "github.com/stv0g/cunicu/pkg/proto/feature/pske"
)

type Peer struct {
	*core.Peer

	Interface *Interface

	SecretKey KyberSecretKey

	logger *zap.Logger
}

func (i *Interface) NewPeer(cp *core.Peer, kem *Interface) *Peer {
	p := &Peer{
		Peer:      cp,
		Interface: i,

		logger: i.logger.With(zap.String("peer", cp.String())),
	}

	go p.EstablishPSK()

	return p
}

func (p *Peer) EstablishPSK() {
	var err error

	ctx := context.Background()
	kp := p.PublicPrivateKeyPair()

	if _, err := p.Interface.Daemon.Backend.Subscribe(ctx, kp, p); err != nil {
		p.logger.Error("Failed to subscribe to message", zap.Error(err))
		return
	}

	if p.IsControlling() {
		var pk KyberPublicKey

		// Generate new key pair
		if p.SecretKey, pk, err = kyber.KemKeypair1024(); err != nil {
			p.logger.Error("Failed to generate Kyber key pair", zap.Error(err))
		}

		// Send public key
		msg := signaling.Message{
			Pske: &pskeproto.PresharedKeyEstablishment{
				PublicKey: pk[:],
			},
		}

		if err := p.Interface.Daemon.Backend.Publish(ctx, kp, &msg); err != nil {
			p.logger.Error("Failed to publish public key", zap.Error(err))
		}
	}
}

func (p *Peer) OnSignalingMessage(kp *crypto.PublicKeyPair, msg *signaling.Message) {
	if msg.Pske == nil {
		return
	}

	var err error
	var psk crypto.Key

	if p.IsControlling() {
		if msg.Pske.CipherText == nil {
			p.logger.Error("Expected cipher text")
			return
		}

		if ctLen := len(msg.Pske.CipherText); ctLen != kyber.Kyber1024CTBytes {
			p.logger.Error("Invalid cipher text length", zap.Int("len", ctLen))
			return
		}

		// Decrypt cipher text
		ct := *(*KyberCipherText)(msg.Pske.CipherText)
		if psk, err = kyber.KemDecrypt1024(ct, p.SecretKey); err != nil {
			p.logger.Error("Failed to decrypt PSK", zap.Error(err))
			return
		}
	} else {
		if msg.Pske.PublicKey == nil {
			p.logger.Error("Expected public key")
			return
		}

		var ct KyberCipherText

		// Encrypt cipher text
		pk := *(*KyberPublicKey)(msg.Pske.PublicKey)
		if ct, psk, err = kyber.KemEncrypt1024(pk); err != nil {
			p.logger.Error("Failed to encrypt cipher text", zap.Error(err))
			return
		}

		// Publish cipher text
		ctx := context.Background()
		kp := p.PublicPrivateKeyPair()
		msg := &signaling.Message{
			Pske: &pskeproto.PresharedKeyEstablishment{
				CipherText: ct[:],
			},
		}

		if err := p.Interface.Daemon.Backend.Publish(ctx, kp, msg); err != nil {
			p.logger.Error("Failed to publish cipher text", zap.Error(err))
		}
	}

	if err := p.SetPresharedKey(&psk); err != nil {
		p.logger.Error("Failed to update preshared key", zap.Error(err))
	}
}
