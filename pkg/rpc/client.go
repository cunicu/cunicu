package rpc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/crypto"
	icex "riasc.eu/wice/pkg/feat/epdisc/ice"
	"riasc.eu/wice/pkg/proto"
	rpcproto "riasc.eu/wice/pkg/proto/rpc"
	"riasc.eu/wice/pkg/util/buildinfo"
)

type Client struct {
	io.Closer

	rpcproto.EndpointDiscoverySocketClient
	rpcproto.SignalingClient
	rpcproto.DaemonClient
	rpcproto.WatcherClient

	conn   *grpc.ClientConn
	logger *zap.Logger

	connectionStates     map[crypto.Key]icex.ConnectionState
	connectionStatesLock sync.Mutex
	connectionStatesCond *sync.Cond

	Events chan *rpcproto.Event
}

func DaemonRunning(path string) bool {
	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: path})
	return err == nil && conn.Close() == nil
}

func waitForSocket(path string) error {
	for tries := 500; tries > 0; tries-- {
		if DaemonRunning(path) {
			return nil
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
	conn, err := grpc.Dial(tgt,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUserAgent(buildinfo.UserAgent()),
	)
	if err != nil {
		return nil, err
	}

	logger := zap.L().Named("rpc.client").With(zap.String("path", path))

	client := &Client{
		EndpointDiscoverySocketClient: rpcproto.NewEndpointDiscoverySocketClient(conn),
		SignalingClient:               rpcproto.NewSignalingClient(conn),
		DaemonClient:                  rpcproto.NewDaemonClient(conn),
		WatcherClient:                 rpcproto.NewWatcherClient(conn),

		conn:             conn,
		logger:           logger,
		connectionStates: make(map[crypto.Key]icex.ConnectionState),
	}
	client.connectionStatesCond = sync.NewCond(&client.connectionStatesLock)

	go client.streamEvents()

	_, err = client.UnWait(context.Background(), &proto.Empty{})
	if sts := status.Convert(err); sts != nil && sts.Code() != codes.AlreadyExists {
		return nil, fmt.Errorf("failed RPC request: %w", err)
	}

	return client, nil
}

func (c *Client) Close() error {
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close gRPC client connection: %w", err)
	}

	// Wait until event channel is closed
	<-c.Events

	return nil
}

func (c *Client) streamEvents() {
	c.Events = make(chan *rpcproto.Event, 100)
	defer close(c.Events)

	stream, err := c.StreamEvents(context.Background(), &proto.Empty{})
	if err != nil {
		c.logger.Error("Failed to stream events", zap.Error(err))
		return
	}

	for {
		e, err := stream.Recv()
		if err != nil {
			if sts, ok := status.FromError(err); !ok || (sts.Code() != codes.Canceled && sts.Code() != codes.Unavailable) {
				c.logger.Error("Failed to receive event", zap.Error(err))
			}

			break
		}

		if e.Type == rpcproto.EventType_PEER_CONNECTION_STATE_CHANGED {
			if pcs, ok := e.Event.(*rpcproto.Event_PeerConnectionStateChange); ok {
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

		c.Events <- e
	}
}

func (c *Client) WaitForEvent(ctx context.Context, t rpcproto.EventType, intf string, peer crypto.Key) (*rpcproto.Event, error) {
	for {
		select {
		case e, ok := <-c.Events:
			if !ok {
				return nil, errors.New("event channel closed")
			}

			if e.Type != t {
				continue
			}

			if intf != "" && intf != e.Interface {
				continue
			}

			if peer.IsSet() && !bytes.Equal(peer.Bytes(), e.Peer) {
				continue
			}

			return e, nil

		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (c *Client) WaitForPeerHandshake(ctx context.Context, peer crypto.Key) error {
	for {
		e, err := c.WaitForEvent(ctx, rpcproto.EventType_PEER_MODIFIED, "", peer)
		if err != nil {
			return err
		}

		ee, ok := e.Event.(*rpcproto.Event_PeerModified)
		if !ok {
			continue
		}

		mod := core.PeerModifier(ee.PeerModified.Modified)
		if mod.Is(core.PeerModifiedHandshakeTime) {
			return nil
		}
	}
}

func (c *Client) WaitForPeerConnectionState(ctx context.Context, peer crypto.Key, csd icex.ConnectionState) error {
	go func() {
		if ch := ctx.Done(); ch != nil {
			<-ch
			c.connectionStatesCond.Broadcast()
		}
	}()

	c.connectionStatesLock.Lock()
	defer c.connectionStatesLock.Unlock()

	for ctx.Err() == nil {
		if cs, ok := c.connectionStates[peer]; ok && cs == csd {
			return nil
		}

		c.connectionStatesCond.Wait()
	}

	return ctx.Err()
}

func (c *Client) RestartPeer(ctx context.Context, intf string, pk *crypto.Key) error {
	_, err := c.EndpointDiscoverySocketClient.RestartPeer(ctx, &rpcproto.RestartPeerParams{
		Intf: intf,
		Peer: pk.Bytes(),
	})

	return err
}
