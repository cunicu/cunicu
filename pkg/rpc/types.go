package rpc

import (
	"fmt"

	"riasc.eu/wice/pkg/pb"
)

type Event struct {
	pb.Event
}

func (e *Event) String() string {
	return fmt.Sprintf("type=%s", e.Type)
}
