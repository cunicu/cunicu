package rpc

import (
	"fmt"

	rpcproto "github.com/stv0g/cunicu/pkg/proto/rpc"
)

type Event struct {
	rpcproto.Event
}

func (e *Event) String() string {
	return fmt.Sprintf("type=%s", e.Type)
}
