package socket

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

type Client struct {
	conn net.Conn

	encoder *json.Encoder
	decoder *json.Decoder

	Events chan Event

	logger *log.Entry
}

func waitForSocket(path string) error {
	tries := 500
	for {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			tries--
			if tries == 0 {
				return fmt.Errorf("timed out")
			} else {
				time.Sleep(10 * time.Millisecond)
			}
			continue
		} else if err != nil {
			return fmt.Errorf("failed stat: %w", err)
		} else {
			break
		}
	}

	return nil
}

func Connect(path string) (*Client, error) {
	if err := waitForSocket(path); err != nil {
		return nil, fmt.Errorf("failed to wait for socket: %w", err)
	}

	conn, err := net.Dial("unix", path)
	if err != nil {
		return nil, err
	}

	logger := log.WithField("logger", "socket")

	client := &Client{
		conn:    conn,
		encoder: json.NewEncoder(conn),
		decoder: json.NewDecoder(conn),

		logger: logger,

		Events: make(chan Event, 100),
	}

	go client.handle()

	return client, nil
}

func (c *Client) handle() {
	for {
		var evt Event

		if err := c.decoder.Decode(&evt); err == io.EOF {
			c.logger.Info("Connection closed")
			break
		} else if err != nil {
			c.logger.WithError(err).Error("Failed to receive event")
		} else {
			evt.Log(c.logger)
			c.Events <- evt
		}
	}
}

func (c *Client) WaitForEvent(flt Event) {
	for evt := range c.Events {
		if flt.Type != "" && flt.Type != evt.Type {
			continue
		}

		if flt.State != "" && flt.State != evt.State {
			continue
		}

		if flt.Interface != "" && flt.Interface != evt.Interface {
			continue
		}

		if flt.Peer.IsSet() && flt.Peer != evt.Peer {
			continue
		}

		return
	}
}

func (c *Client) WaitPeerHandshake() {
	c.WaitForEvent(Event{
		Type: "handshake",
	})
}

func (c *Client) WaitPeerConnected() {
	c.WaitForEvent(Event{
		Type:  "state",
		State: "connected",
	})
}
