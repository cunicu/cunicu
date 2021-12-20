package socket

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"riasc.eu/wice/pkg/crypto"
)

type Request struct {
	Type string
}

type Response struct {
}

type Event struct {
	Type string `json:"type"`

	State string `json:"state"`

	Interface string
	Peer      crypto.Key

	Time time.Time `json:"time"`
}

func (e *Event) String() string {
	return fmt.Sprintf("type=%s", e.Type)
}

func (r *Request) String() string {
	return fmt.Sprintf("type=%s", r.Type)
}

func (e *Event) Log(logger *log.Entry) {
	fields := log.Fields{
		"type":  e.Type,
		"state": e.State,
	}

	if e.Interface != "" {
		fields["intf"] = e.Interface
	}

	if e.Peer.IsSet() {
		fields["peer"] = e.Peer.String()
	}

	logger.WithFields(fields).Infof("Received event")
}
