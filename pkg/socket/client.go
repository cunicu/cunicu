package socket

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/pion/ice/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	ginsecure "google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
)

type Client struct {
	io.Closer

	pb.SocketClient
	grpc   *grpc.ClientConn
	logger *zap.Logger

	connectionStates     map[crypto.Key]ice.ConnectionState
	connectionStatesLock sync.Mutex
	connectionStatesCond *sync.Cond

	Events chan *pb.Event
}

func waitForSocket(path string) error {
	for tries := 500; tries > 0; tries-- {
		ua := &net.UnixAddr{Name: path}
		if conn, err := net.DialUnix("unix", nil, ua); err == nil {
			return conn.Close()
		}

		time.Sleep(10 * time.Millisecond)
	}

	return fmt.Errorf("timed out")
}

func Connect(path string) (*Client, error) {
	if err := waitForSocket(path); err != nil {
		return nil, fmt.Errorf("failed to wait for socket: %s: %w", path, err)
	}

	tgt := fmt.Sprintf("unix://%s", path)
	conn, err := grpc.Dial(tgt, grpc.WithTransportCredentials(ginsecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	logger := zap.L().Named("socket.client").With(zap.String("path", path))

	client := &Client{
		SocketClient:     pb.NewSocketClient(conn),
		grpc:             conn,
		logger:           logger,
		Events:           make(chan *pb.Event, 100),
		connectionStates: make(map[crypto.Key]ice.ConnectionState),
	}
	client.connectionStatesCond = sync.NewCond(&client.connectionStatesLock)

	go client.streamEvents()

	rerr, err := client.UnWait(context.Background(), &pb.UnWaitParams{})
	if err != nil {
		return nil, fmt.Errorf("failed RPC request: %w", err)
	} else if !rerr.Ok() && rerr.Code != pb.Error_EALREADY {
		return nil, fmt.Errorf("received RPC error: %w", rerr)
	}

	return client, nil
}

func (c *Client) Close() error {
	close(c.Events)

	return c.grpc.Close()
}

func (c *Client) streamEvents() {
	stream, err := c.StreamEvents(context.Background(), &pb.StreamEventsParams{})
	if err != nil {
		c.logger.Error("Failed to stream events", zap.Error(err))
		return
	}

	ok := true
	for ok {
		e, err := stream.Recv()
		if err != nil {
			sts, ok := status.FromError(err)
			if !ok || (sts.Code() != codes.Canceled && sts.Code() != codes.Unavailable) {
				c.logger.Error("Failed to receive event", zap.Error(err))
			}

			break
		}

		if e.Type == pb.Event_PEER_CONNECTION_STATE_CHANGED {
			if pcs, ok := e.Event.(*pb.Event_PeerConnectionStateChange); ok {
				pk, err := crypto.ParseKeyBytes(e.Peer)
				if err != nil {
					c.logger.Error("Invalid key", zap.Error(err))
					continue
				}

				cs := pcs.PeerConnectionStateChange.NewState.ConnectionState()

				c.connectionStatesLock.Lock()
				c.connectionStates[pk] = cs
				c.connectionStatesCond.Broadcast()
				c.connectionStatesLock.Unlock()
			}
		}

		e.Log(c.logger, "Received event")
		c.Events <- e
	}
}

func (c *Client) WaitForEvent(t pb.Event_Type, intf string, peer crypto.Key) *pb.Event {
	for e := range c.Events {
		if e.Type != t {
			continue
		}

		if intf != "" && intf != e.Interface {
			continue
		}

		if peer.IsSet() && !bytes.Equal(peer.Bytes(), e.Peer) {
			continue
		}

		return e
	}

	return nil
}

func (c *Client) WaitForPeerHandshake(peer crypto.Key) {
	for {
		e := c.WaitForEvent(pb.Event_PEER_MODIFIED, "", peer)

		ee, ok := e.Event.(*pb.Event_PeerModified)
		if !ok {
			continue
		}

		mod := core.PeerModifier(ee.PeerModified.Modified)
		if mod.Is(core.PeerModifiedHandshakeTime) {
			return
		}
	}
}

func (c *Client) WaitForPeerConnectionState(peer crypto.Key, csd ice.ConnectionState) {
	for {
		c.connectionStatesLock.Lock()
		for {
			if cs, ok := c.connectionStates[peer]; ok && cs == csd {
				c.connectionStatesLock.Unlock()
				return
			}

			c.connectionStatesCond.Wait()
		}
	}
}
