package ep

import "github.com/pion/ice/v2"

type OnConnectionStateHandler interface {
	OnConnectionStateChange(*Peer, ice.ConnectionState)
}
