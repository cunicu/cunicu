// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package epdisc

import (
	epdiscproto "github.com/stv0g/cunicu/pkg/proto/feature/epdisc"
)

type ConnectionState = epdiscproto.ConnectionState

const (
	ConnectionStateNew          = epdiscproto.ConnectionState_NEW
	ConnectionStateChecking     = epdiscproto.ConnectionState_CHECKING
	ConnectionStateConnected    = epdiscproto.ConnectionState_CONNECTED
	ConnectionStateCompleted    = epdiscproto.ConnectionState_COMPLETED
	ConnectionStateFailed       = epdiscproto.ConnectionState_FAILED
	ConnectionStateDisconnected = epdiscproto.ConnectionState_DISCONNECTED
	ConnectionStateClosed       = epdiscproto.ConnectionState_CLOSED
)

// The following connection states are an extension to the states by the ICE RFC
// in order to mitigate race conditions when handling the pion/ice.Agent.
// They are mainly used for transitioning between the states above.
const (
	ConnectionStateConnecting ConnectionState = 100 + iota
	ConnectionStateClosing
	ConnectionStateCreating
	ConnectionStateRestarting
	ConnectionStateIdle
	ConnectionStateGathering
	ConnectionStateGatheringLocal  // After first remote candidate has been received
	ConnectionStateGatheringRemote // After first local candidate has been received
)
