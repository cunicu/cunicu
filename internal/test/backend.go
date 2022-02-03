package test

import (
	"net/url"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"riasc.eu/wice/internal/log"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/signaling"
)

func TestBackend(t *testing.T, u string) {
	if !strings.Contains(u, ":") {
		u += ":"
	}

	uri, err := url.Parse(u)
	if err != nil {
		t.Fatalf("Failed to parse URL: %s", err)
	}

	com := crypto.GenerateKeyFromPassword("")

	cfg := &signaling.BackendConfig{
		URI:       uri,
		Community: &com,
	}

	events := log.NewEventLogger()

	ourBackend, err := signaling.NewBackend(cfg, events)
	if err != nil {
		t.Fatalf("Failed to create backend: %s", err)
	}
	defer ourBackend.Close()

	theirBackend, err := signaling.NewBackend(cfg, events)
	if err != nil {
		t.Fatalf("Failed to create backend: %s", err)
	}
	defer theirBackend.Close()

	sentMsg := GenerateSignalingMessage()

	ourKP, theirKP, err := GenerateKeyPairs()
	if err != nil {
		t.Fatalf("Failed to generate keypairs: %s", err)
	}

	ch, err := theirBackend.Subscribe(theirKP)
	if err != nil {
		t.Fatalf("Failed to subscribe to signaling messages: %s", err)
	}

	if err := ourBackend.Publish(ourKP, sentMsg); err != nil {
		t.Fatalf("Failed to publish signaling message: %s", err)
	}

	recvMsg := <-ch

	if !proto.Equal(recvMsg, sentMsg) {
		t.Fatalf("Sent and received messages are not equal!\n%#+v\n%#+v", recvMsg, sentMsg)
	}
}
