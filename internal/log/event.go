package log

import (
	"go.uber.org/zap"
	"riasc.eu/wice/pkg/pb"
)

func NewEventLogger() chan *pb.Event {
	events := make(chan *pb.Event)

	go drainEvents(events)

	return events
}

func drainEvents(events chan *pb.Event) {
	logger := zap.L().Named("events")

	for event := range events {
		event.Log(logger, "New event")
	}
}
