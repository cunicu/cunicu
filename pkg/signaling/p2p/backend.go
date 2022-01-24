package p2p

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	p2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	noise "github.com/libp2p/go-libp2p-noise"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	ptls "github.com/libp2p/go-libp2p-tls"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/host/autorelay"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling"
)

const (
	// See: https://github.com/multiformats/multicodec/blob/master/table.csv#L85
	CodeX25519PublicKey = 0xec

	userAgent = "wice"
)

func init() {
	signaling.Backends["p2p"] = &signaling.BackendPlugin{
		New:         NewBackend,
		Description: "libp2p",
	}
}

type Backend struct {
	logger *zap.Logger
	config BackendConfig

	peers     map[crypto.KeyPair]*Peer
	peersLock sync.Mutex

	host host.Host

	context context.Context

	mdns   mdns.Service
	dht    *dht.IpfsDHT
	pubsub *pubsub.PubSub
	events chan *pb.Event
}

func NewBackend(uri *url.URL, events chan *pb.Event) (signaling.Backend, error) {
	var err error

	b := &Backend{
		peers:  map[crypto.KeyPair]*Peer{},
		logger: zap.L().Named("backend").With(zap.String("backend", uri.Scheme)),
		config: defaultConfig,
		events: events,
	}

	if err := b.config.Parse(uri); err != nil {
		return nil, fmt.Errorf("failed to parse backend options: %w", err)
	}

	b.context = context.Background()

	opts := []p2p.Option{
		p2p.UserAgent(userAgent),
		p2p.DefaultTransports,
		p2p.EnableNATService(),
		p2p.Security(ptls.ID, ptls.New),
		p2p.Security(noise.ID, noise.New),
	}

	if len(b.config.ListenAddresses) > 0 {
		opts = append(opts, p2p.ListenAddrs([]multiaddr.Multiaddr(b.config.ListenAddresses)...))
	} else {
		opts = append(opts, p2p.DefaultListenAddrs)
	}

	if !b.config.EnableRelay {
		opts = append(opts, p2p.DisableRelay())
	}

	if b.config.EnableAutoRelay {
		opts = append(opts, p2p.EnableAutoRelay(
			autorelay.WithStaticRelays(b.config.AutoRelayPeers),
		))
	}

	if b.config.EnableNATPortMap {
		opts = append(opts, p2p.NATPortMap())
	}

	if b.config.EnableHolePunching {
		opts = append(opts, p2p.EnableHolePunching())
	}

	if b.config.PrivateKey != nil {
		opts = append(opts, p2p.Identity(b.config.PrivateKey))
	}

	// if b.config.EnableDHTDiscovery {
	// 	opts = append(opts, p2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
	// 		b.dht, err = dht.New(b.context, h)
	// 		return b.dht, err
	// 	}))
	// }

	// create host
	if b.host, err = p2p.New(opts...); err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}
	b.logger.Info("Host created",
		zap.Any("id", b.host.ID()),
		zap.Any("addrs", b.host.Addrs()),
	)

	b.host.Network().Notify(newNotifee(b))

	// setup local mDNS discovery
	if b.config.EnableMDNSDiscovery {
		b.logger.Debug("Setup mDNS discovery")

		b.mdns = mdns.NewMdnsService(b.host, b.config.MDNSServiceTag, b)
		if err := b.mdns.Start(); err != nil {
			return nil, fmt.Errorf("failed to start mDNS service: %w", err)
		}
	}

	// setup DHT discovery
	if b.config.EnableDHTDiscovery {
		b.logger.Debug("Bootstrapping the DHT")

		if b.dht, err = dht.New(b.context, b.host,
			dht.Mode(dht.ModeServer),
		); err != nil {
			return nil, fmt.Errorf("failed to create DHT: %w", err)
		}

		if err = b.dht.Bootstrap(b.context); err != nil {
			return nil, fmt.Errorf("failed to bootstrap DHT: %w", err)
		}
	}

	rt := b.dht.RoutingTable()

	// Add some handlers
	rt.PeerAdded = func(i peer.ID) {
		b.logger.Debug("Peer added to routing table", zap.Any("peer", i))
	}

	rt.PeerRemoved = func(i peer.ID) {
		b.logger.Debug("Peer removed from routing table", zap.Any("peer", i))
	}

	// Let's connect to the bootstrap nodes first. They will tell us about the
	// other nodes in the network.
	var wg sync.WaitGroup
	for _, pi := range b.config.BootstrapPeers {
		logger := b.logger.With(zap.Any("peer", pi))

		logger.Debug("Connecting to peer")
		wg.Add(1)
		go func(pi peer.AddrInfo) {
			defer wg.Done()

			if err := b.host.Connect(b.context, pi); err != nil {
				logger.Warn("Failed to connect to boostrap node")
			} else {
				logger.Info("Connection established with bootstrap node")
			}
		}(pi)
	}
	wg.Wait() // TODO: can we run this asynchronously?

	rd := discovery.NewRoutingDiscovery(b.dht)

	// setup PubSub service using the GossipSub router
	if b.pubsub, err = pubsub.NewGossipSub(b.context, b.host,
		pubsub.WithDiscovery(rd),
		pubsub.WithRawTracer(newTracer(b)),
	); err != nil {
		return nil, fmt.Errorf("failed to create pubsub router: %w", err)
	}

	as := []string{}
	for _, a := range b.host.Addrs() {
		as = append(as, a.String())
	}

	b.events <- &pb.Event{
		Type: pb.Event_BACKEND_READY,
		Event: &pb.Event_BackendReady{
			BackendReady: &pb.BackendReadyEvent{
				Type:            pb.BackendReadyEvent_P2P,
				Id:              b.host.ID().String(),
				ListenAddresses: as,
			},
		},
	}

	return b, nil
}

func (b *Backend) getPeer(kp crypto.KeyPair) (*Peer, error) {
	var err error

	b.peersLock.Lock()
	defer b.peersLock.Unlock()

	p, ok := b.peers[kp]
	if !ok {
		if p, err = b.NewPeer(kp); err != nil {
			return nil, err
		}

		b.peers[kp] = p
	}

	return p, nil
}

func (b *Backend) SubscribeOffers(kp crypto.KeyPair) (chan *pb.Offer, error) {
	b.logger.Info("Subscribe to offers from peer", zap.Any("kp", kp))

	p, err := b.getPeer(kp)
	if err != nil {
		return nil, err
	}

	return p.Offers, nil
}

func (b *Backend) PublishOffer(kp crypto.KeyPair, offer *pb.Offer) error {
	p, err := b.getPeer(kp)
	if err != nil {
		return fmt.Errorf("failed to get peer: %w", err)
	}

	if err := p.publishOffer(offer); err != nil {
		return err
	}

	return nil
}

func (b *Backend) Close() error {
	return nil // TODO
}

func (b *Backend) Tick() {}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (b *Backend) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == b.host.ID() {
		return // skip ourself
	}

	b.logger.Info("Discovered new peer via mDNS",
		zap.Any("peer", pi.ID))

	if err := b.host.Connect(b.context, pi); err != nil {
		b.logger.Error("Failed connecting to peer",
			zap.Any("peer", pi.ID),
			zap.Error(err),
		)
	}
}
