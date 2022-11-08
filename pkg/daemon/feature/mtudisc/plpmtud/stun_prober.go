package plpmtud

import (
	"errors"
	"fmt"
	"net"

	"github.com/pion/stun"
	"go.uber.org/zap"
)

// A PLPMTUD prober for UDP transports using STUN
// See the following RFC draft: https://datatracker.ietf.org/doc/html/draft-petithuguenin-tsvwg-stun-pmtud-01
type StunProber struct {
	*StunMultiplexer

	discoverer *Discoverer

	localUfrag  string
	localPwd    string
	remoteUfrag string
	remotePwd   string

	logger *zap.Logger
}

func NewStunConnProber(c *net.UDPConn) (*StunProber, error) {
	m, err := NewStunMultiplexer(c)
	if err != nil {
		return nil, err
	}

	return NewStunMultiplexProber(m)
}

func NewStunMultiplexProber(m *StunMultiplexer) (*StunProber, error) {
	sp := &StunProber{
		StunMultiplexer: m,
		logger:          zap.L().Named("stun-prober"),
	}

	sp.RegisterStunHandler(StunMethodProbe, sp.onProbe)

	return sp, nil
}

func (p *StunProber) Close() error {
	return nil
}

func (p *StunProber) SendProbeRequest(mtu uint) error {
	if p.localPwd == "" || p.localUfrag == "" {
		return errors.New("no credentials set")
	}

	paddingLength := mtu
	padding := make([]byte, paddingLength)

	msg := stun.New()

	stun.NewUsername(p.localUfrag).AddTo(msg)
	stun.NewShortTermIntegrity(p.localPwd).AddTo(msg)
	msg.Add(StunAttrPadding, padding)
	stun.Fingerprint.AddTo(msg)

	if err := p.WriteStunMessage(msg); err != nil {
		return fmt.Errorf("failed to send probe message: %w", err)
	}

	return nil
}

func (p *StunProber) SendProbeResponse(mtu uint) error {
	msg := stun.New()

	stun.NewUsername(p.localUfrag).AddTo(msg)
	stun.NewShortTermIntegrity(p.localPwd).AddTo(msg)

	if err := p.WriteStunMessage(msg); err != nil {
		return fmt.Errorf("failed to send probe message: %w", err)
	}

	return nil
}

func (p *StunProber) RegisterDiscoverer(h *Discoverer) {
	p.discoverer = h
}

func (p *StunProber) SetCredentials(localUfrag, localPwd, remoteUfrag, remotePwd string) {
	p.localUfrag = localUfrag
	p.localPwd = localPwd

	p.remoteUfrag = remoteUfrag
	p.remotePwd = remotePwd
}

func (p *StunProber) onProbe(msg *stun.Message) error {
	if p.discoverer == nil {
		return nil
	}

	if msg.Type.Method != StunMethodProbe {
		return fmt.Errorf("expected probe")
	}

	usernameAttr := stun.Username{}
	if err := usernameAttr.GetFrom(msg); err != nil {
		return fmt.Errorf("missing username attribute")
	}

	if string(usernameAttr) != p.remoteUfrag {
		return fmt.Errorf("mismatching ufrag: expected=%s, received=%s", p.remotePwd, usernameAttr)
	}

	integrityAttr := stun.NewShortTermIntegrity(p.remotePwd)
	if err := integrityAttr.Check(msg); err != nil {
		return fmt.Errorf("invalid message integrity: %w", err)
	}

	switch msg.Type.Class {
	case stun.ClassSuccessResponse:
		p.discoverer.OnProbeResponse(0)

	case stun.ClassRequest:
		p.discoverer.OnProbeRequest(0)
	}

	return nil
}
