// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package daemon

import (
	"strings"
	"time"

	"go.uber.org/zap"

	coreproto "github.com/stv0g/cunicu/pkg/proto/core"
)

type PeerState = coreproto.PeerState

// Prettier aliases for the protobuf constants
const (
	PeerStateNew        = coreproto.PeerState_NEW
	PeerStateConnecting = coreproto.PeerState_CONNECTING
	PeerStateConnected  = coreproto.PeerState_CONNECTED
	PeerStateFailed     = coreproto.PeerState_FAILED
	PeerStateClosed     = coreproto.PeerState_CLOSED
)

func (p *Peer) State() PeerState {
	return p.state.Load()
}

// SetStateIf updates the connection state of the peer if the previous state
// matches one of the supplied previous states.
// It returns true if the state has been changed.
func (p *Peer) SetStateIf(newState PeerState, prevStates ...PeerState) (PeerState, bool) {
	prevState, ok := p.state.SetIf(newState, prevStates...)
	if ok {
		p.onStateChanged(newState, prevState)
	}

	return prevState, ok
}

// SetStateIf updates the connection state of the peer if the previous state
// does not match any of the supplied previous states.
func (p *Peer) SetStateIfNot(newState PeerState, prevStates ...PeerState) (PeerState, bool) {
	prevState, ok := p.state.SetIfNot(newState, prevStates...)
	if ok {
		p.onStateChanged(newState, prevState)
	}

	return prevState, ok
}

// onStateChanged emits a log message about the changed state and call registered handlers
func (p *Peer) onStateChanged(newState, prevState PeerState) {
	p.LastStateChangeTime = time.Now()

	p.logger.Info("State changed",
		zap.String("new", strings.ToLower(newState.String())),
		zap.String("previous", strings.ToLower(prevState.String())))

	for _, h := range p.Interface.onPeerStateChanged {
		h.OnPeerStateChanged(p, newState, prevState)
	}
}
