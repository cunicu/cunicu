package rpc

import (
	"fmt"

	rpcproto "riasc.eu/wice/pkg/proto/rpc"
)

type Event struct {
	rpcproto.Event
}

func (e *Event) String() string {
	return fmt.Sprintf("type=%s", e.Type)
}
