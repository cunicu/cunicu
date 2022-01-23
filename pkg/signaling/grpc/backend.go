package grpc

import (
	"context"
	"fmt"
	"net/url"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling"
)

func init() {
	signaling.Backends["grpc"] = &signaling.BackendPlugin{
		New:         NewBackend,
		Description: "gRPC",
	}
}

type Backend struct {
	client pb.SignalingClient
	conn   *grpc.ClientConn

	config BackendConfig

	logger *zap.Logger
}

func NewBackend(uri *url.URL, events chan *pb.Event) (signaling.Backend, error) {
	var err error

	b := &Backend{}

	if err := b.config.Parse(uri); err != nil {
		return nil, fmt.Errorf("failed to parse backend configuration: %w", err)
	}

	if b.conn, err = grpc.Dial(b.config.Target, b.config.Options...); err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	b.client = pb.NewSignalingClient(b.conn)

	return b, nil
}

func (b *Backend) SubscribeOffer(kp crypto.PublicKeyPair) (chan *pb.Offer, error) {
	params := &pb.SubscribeOffersParams{
		SharedKey: kp.Shared().Bytes(),
	}

	stream, err := b.client.SubscribeOffers(context.Background(), params)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to offers: %s", err)
	}

	ch := make(chan *pb.Offer)

	go func() {
		for {
			if offer, err := stream.Recv(); err == nil {
				ch <- offer
			} else {
				b.logger.Error("Failed to receive offer", zap.Error(err))
			}
		}
	}()

	return ch, nil
}

func (b *Backend) PublishOffer(kp crypto.PublicKeyPair, offer *pb.Offer) error {
	params := &pb.PublishOffersParams{
		SharedKey: kp.Shared().Bytes(),
		Offer:     offer,
	}

	_, err := b.client.PublishOffer(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to publish offer: %w", err)
	}

	return nil
}

func (b *Backend) Close() error {
	if err := b.conn.Close(); err != nil {
		return fmt.Errorf("failed to close gRPC connection: %w", err)
	}

	return nil
}

func (b *Backend) Tick() {}
