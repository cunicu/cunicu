// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package rpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/stv0g/cunicu/pkg/buildinfo"
	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/log"
	"github.com/stv0g/cunicu/pkg/proto"
	rpcproto "github.com/stv0g/cunicu/pkg/proto/rpc"
)

var errTimeout = errors.New("timed out")

type EventHandler interface {
	OnEvent(*rpcproto.Event)
}

type Client struct {
	io.Closer

	rpcproto.EndpointDiscoverySocketClient
	rpcproto.SignalingClient
	rpcproto.DaemonClient

	conn    *grpc.ClientConn
	logger  *log.Logger
	onEvent []EventHandler

	peerStates     map[crypto.Key]daemon.PeerState
	peerStatesLock sync.Mutex
	peerStatesCond *sync.Cond
}

func DaemonRunning(path string) bool {
	conn, err := net.Dial("unix", path)
	return err == nil && conn.Close() == nil
}

func waitForSocket(path string) error {
	for tries := 200; tries > 0; tries-- {
		if DaemonRunning(path) {
			return nil
		}

		time.Sleep(10 * time.Millisecond)
	}

	return errTimeout
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

	c := &Client{
		EndpointDiscoverySocketClient: rpcproto.NewEndpointDiscoverySocketClient(conn),
		SignalingClient:               rpcproto.NewSignalingClient(conn),
		DaemonClient:                  rpcproto.NewDaemonClient(conn),

		conn:       conn,
		logger:     log.Global.Named("rpc.client").With(zap.String("path", path)),
		peerStates: make(map[crypto.Key]daemon.PeerState),
	}
	c.peerStatesCond = sync.NewCond(&c.peerStatesLock)

	c.AddEventHandler(c)

	go c.streamEvents()

	return c, nil
}

func (c *Client) Close() error {
	if err := c.conn.Close(); err != nil && status.Code(err) != codes.Canceled {
		return fmt.Errorf("failed to close gRPC client connection: %w", err)
	}

	c.logger.Debug("Closed")

	return nil
}

func (c *Client) streamEvents() {
	stream, err := c.StreamEvents(context.Background(), &proto.Empty{})
	if err != nil {
		c.logger.Error("Failed to stream events", zap.Error(err))
		return
	}

	for {
		e, err := stream.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) && status.Code(err) != codes.Canceled {
				c.logger.Error("Failed to receive event", zap.Error(err))
			}

			break
		}

		for _, h := range c.onEvent {
			h.OnEvent(e)
		}
	}
}

type waitHandler struct {
	event chan *rpcproto.Event

	typ  rpcproto.EventType
	intf string
	peer crypto.Key
}

func (h *waitHandler) OnEvent(e *rpcproto.Event) {
	peer, err := crypto.ParseKeyBytes(e.Peer)
	if err != nil {
		panic(err)
	}

	if (e.Type != h.typ) ||
		(h.intf != "" && h.intf != e.Interface) ||
		(h.peer.IsSet() && h.peer != peer) {
		return
	}
}

func (c *Client) WaitForEvent(ctx context.Context, t rpcproto.EventType, intf string, peer crypto.Key) (*rpcproto.Event, error) {
	h := &waitHandler{
		event: make(chan *rpcproto.Event),

		typ:  t,
		intf: intf,
		peer: peer,
	}

	c.AddEventHandler(h)
	defer c.RemoveEventHandler(h)
	defer close(h.event)

	for {
		select {
		case e := <-h.event:
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

		mod := daemon.PeerModifier(ee.PeerModified.Modified)
		if mod.Is(daemon.PeerModifiedHandshakeTime) {
			return nil
		}
	}
}

func (c *Client) WaitForPeerState(ctx context.Context, peer crypto.Key, csd daemon.PeerState) error {
	go func() {
		if ch := ctx.Done(); ch != nil {
			<-ch
			c.peerStatesCond.Broadcast()
		}
	}()

	c.peerStatesLock.Lock()
	defer c.peerStatesLock.Unlock()

	for ctx.Err() == nil {
		if cs, ok := c.peerStates[peer]; ok && cs == csd {
			return nil
		}

		c.peerStatesCond.Wait()
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

func (c *Client) Unwait() error {
	_, err := c.DaemonClient.UnWait(context.Background(), &proto.Empty{})
	if sts := status.Convert(err); sts != nil && sts.Code() != codes.AlreadyExists {
		return fmt.Errorf("failed RPC request: %w", err)
	}

	return nil
}

func (c *Client) OnEvent(e *rpcproto.Event) {
	if e.Type == rpcproto.EventType_PEER_STATE_CHANGED {
		if psc, ok := e.Event.(*rpcproto.Event_PeerStateChange); ok {
			pk, err := crypto.ParseKeyBytes(e.Peer)
			if err != nil {
				c.logger.Error("Invalid key", zap.Error(err))
				return
			}

			c.peerStatesLock.Lock()
			c.peerStates[pk] = psc.PeerStateChange.NewState
			c.peerStatesCond.Broadcast()
			c.peerStatesLock.Unlock()
		}
	}
}

func (c *Client) AddEventHandler(h EventHandler) {
	if !slices.Contains(c.onEvent, h) {
		c.onEvent = append(c.onEvent, h)
	}
}

func (c *Client) RemoveEventHandler(h EventHandler) {
	if idx := slices.Index(c.onEvent, h); idx > -1 {
		c.onEvent = slices.Delete(c.onEvent, idx, idx+1)
	}
}
