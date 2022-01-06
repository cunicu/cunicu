package socket

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	ginsecure "google.golang.org/grpc/credentials/insecure"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
)

type Client struct {
	io.Closer

	pb.SocketClient
	grpc   *grpc.ClientConn
	logger *zap.Logger

	Events chan *pb.Event
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

	tgt := fmt.Sprintf("unix://%s", path)
	conn, err := grpc.Dial(tgt, grpc.WithTransportCredentials(ginsecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	logger := zap.L().Named("socket.client").With(zap.String("path", path))

	client := &Client{
		SocketClient: pb.NewSocketClient(conn),
		grpc:         conn,
		logger:       logger,
		Events:       make(chan *pb.Event, 100),
	}

	rerr, err := client.UnWait(context.Background(), &pb.UnWaitParams{})
	if err != nil {
		return nil, fmt.Errorf("failed RPC request: %w", err)
	} else if !rerr.Ok() && rerr.Code != pb.Error_EALREADY {
		return nil, fmt.Errorf("received RPC error: %w", rerr)
	}

	go client.streamEvents()

	return client, nil
}

func (c *Client) Close() error {
	close(c.Events)

	return c.grpc.Close()
}

func (c *Client) streamEvents() {
	str, err := c.StreamEvents(context.Background(), &pb.StreamEventsParams{})
	if err != nil {
		c.logger.Error("Failed to stream events", zap.Error(err))
	}

	ok := true
	for ok {
		evt, err := str.Recv()
		if err != nil {
			c.logger.Error("Failed to receive event", zap.Error(err))
			break
		}

		evt.Log(c.logger, "Received event")
		c.Events <- evt
	}
}

func (c *Client) WaitForEvent(flt *pb.Event) *pb.Event {
	for evt := range c.Events {
		if evt.Match(flt) {
			return evt
		}
	}

	return nil
}

func (c *Client) WaitPeerHandshake(peer crypto.Key) {
	c.WaitForEvent(&pb.Event{
		Type:  "handshake",
		State: "new",
	})
}

func (c *Client) WaitPeerConnected() {
	c.WaitForEvent(&pb.Event{
		Type:  "state",
		State: "connected",
	})
}
