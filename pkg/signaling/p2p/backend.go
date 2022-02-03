package p2p

import (
	"context"
	"fmt"
	"net"
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
	ma "github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
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
	signaling.SubscriptionsRegistry

	logger *zap.Logger
	config BackendConfig

	context context.Context

	host host.Host
	mdns mdns.Service

	dht    *dht.IpfsDHT
	pubsub *pubsub.PubSub
	topic  *pubsub.Topic
	sub    *pubsub.Subscription

	events chan *pb.Event
}

func NewBackend(cfg *signaling.BackendConfig, events chan *pb.Event) (signaling.Backend, error) {
	var err error

	b := &Backend{
		SubscriptionsRegistry: signaling.NewSubscriptionsRegistry(),

		logger: zap.L().Named("backend").With(zap.String("backend", cfg.URI.Scheme)),
		config: defaultConfig,
		events: events,
	}

	if err := b.config.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse backend options: %w", err)
	}

	b.logger.Debug("Config", zap.Any("config", b.config))

	b.context = context.Background()

	opts := b.options()
	if err = b.setupHost(opts...); err != nil {
		return nil, err
	}

	if b.config.EnableDHTDiscovery {
		if err := b.setupDHT(); err != nil {
			return nil, fmt.Errorf("failed to setup DHT: %w", err)
		}
	}

	if b.config.EnableMDNSDiscovery {
		if err := b.setupMDNS(); err != nil {
			return nil, fmt.Errorf("failed to setup MDNS: %w", err)
		}
	}

	if !b.config.Private {
		b.bootstrap()
	}

	if err := b.setupPubSub(); err != nil {
		return nil, fmt.Errorf("failed to setup pubsub: %w", err)
	}

	b.logger.Info("Node libp2p adresses", zap.Strings("addresses", b.StringAddrs()))

	b.events <- &pb.Event{
		Type: pb.Event_BACKEND_READY,
		Event: &pb.Event_BackendReady{
			BackendReady: &pb.BackendReadyEvent{
				Type:            pb.BackendReadyEvent_P2P,
				Id:              b.host.ID().String(),
				ListenAddresses: b.StringAddrs(),
			},
		},
	}

	return b, nil
}

func (b *Backend) StringAddrs() []string {
	p2p, err := ma.NewMultiaddr(fmt.Sprintf("/p2p/%s", b.host.ID()))
	if err != nil {
		return nil
	}

	as := []string{}
out:
	for _, a := range b.host.Addrs() {
		// Skip IPv4 and IPv6 loopback addresses
		for _, prot := range []int{ma.P_IP4, ma.P_IP6} {
			if val, err := a.ValueForProtocol(prot); err == nil {
				if ip := net.ParseIP(val); ip.IsLoopback() {
					continue out
				}
			}
		}

		a = a.Encapsulate(p2p)
		as = append(as, a.String())
	}
	return as
}

func (b *Backend) options() []p2p.Option {
	opts := []p2p.Option{
		p2p.UserAgent(userAgent),
		p2p.DefaultTransports,
		p2p.EnableNATService(),
		p2p.EnableRelayService(),
		p2p.Security(ptls.ID, ptls.New),
		p2p.Security(noise.ID, noise.New),
	}

	if len(b.config.ListenAddresses) > 0 {
		opts = append(opts, p2p.ListenAddrs([]ma.Multiaddr(b.config.ListenAddresses)...))
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

	if b.config.PrivateCommunity {
		opts = append(opts, p2p.PrivateNetwork(b.config.Community[:]))
	}

	// if b.config.EnableDHTDiscovery {
	// 	opts = append(opts, p2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
	// 		b.dht, err = dht.New(b.context, h)
	// 		return b.dht, err
	// 	}))
	// }

	return opts
}

func (b *Backend) setupHost(opts ...p2p.Option) error {
	var err error

	b.host, err = p2p.New(opts...)

	if err != nil {
		return fmt.Errorf("failed to create host: %w", err)
	}
	b.logger.Info("Host created",
		zap.Any("id", b.host.ID()),
		zap.Any("addrs", b.host.Addrs()),
	)

	// Setup logging and tracing
	b.host.Network().Notify(newNotifee(b))

	return nil
}

func (b *Backend) setupDHT() error {
	var err error

	b.logger.Debug("Bootstrapping the DHT")

	if b.dht, err = dht.New(b.context, b.host,
		dht.Mode(dht.ModeServer),
	); err != nil {
		return fmt.Errorf("failed to create DHT: %w", err)
	}

	rt := b.dht.RoutingTable()

	// Register some event handlers
	rt.PeerAdded = func(i peer.ID) {
		b.logger.Debug("Peer added to routing table", zap.Any("peer", i))
	}

	rt.PeerRemoved = func(i peer.ID) {
		b.logger.Debug("Peer removed from routing table", zap.Any("peer", i))
	}

	if err = b.dht.Bootstrap(b.context); err != nil {
		return fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	return nil
}

func (b *Backend) setupMDNS() error {
	b.logger.Debug("Setup mDNS discovery")

	b.mdns = mdns.NewMdnsService(b.host, b.config.MDNSServiceTag, b)
	if err := b.mdns.Start(); err != nil {
		return fmt.Errorf("failed to start mDNS service: %w", err)
	}

	return nil
}

func (b *Backend) setupPubSub() error {
	var err error

	opts := []pubsub.Option{
		pubsub.WithRawTracer(newTracer(b)),
	}

	if b.dht != nil {
		rd := discovery.NewRoutingDiscovery(b.dht)
		opts = append(opts, pubsub.WithDiscovery(rd))
	}

	// Setup PubSub service using the GossipSub router
	if b.pubsub, err = pubsub.NewGossipSub(b.context, b.host, opts...); err != nil {
		return fmt.Errorf("failed to create pubsub router: %w", err)
	}

	// Setup pubsub topic subscription
	t := fmt.Sprintf("wice/%s", b.config.Community)

	if b.topic, err = b.pubsub.Join(t); err != nil {
		return fmt.Errorf("failed to join topic: %w", err)
	}

	if b.sub, err = b.topic.Subscribe(); err != nil {
		return fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	go b.subscriberLoop()

	return nil
}

func (b *Backend) bootstrap() {
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
}

func (b *Backend) subscriberLoop() {
	for {
		msg, err := b.sub.Next(b.context)
		if err != nil {
			b.logger.Error("Failed to receive offers", zap.Error(err))
			return
		}

		// Skip our own data
		if msg.ReceivedFrom == b.host.ID() {
			continue
		}

		env := &pb.SignalingEnvelope{}
		if err := proto.Unmarshal(msg.Data, env); err != nil {
			b.logger.Error("Failed to unmarshal offer", zap.Error(err))
			continue
		}

		sender, err := crypto.ParseKeyBytes(env.Sender)
		if err != nil {
			b.logger.Error("Invalid key", zap.Error(err))
			continue
		}

		sub, err := b.GetSubscription(&sender)
		if err != nil {
			b.logger.Error("Failed to get subscription", zap.Error(err))
			continue
		}

		sub.NewMessage(env)
	}
}

func (b *Backend) Subscribe(kp *crypto.KeyPair) (chan *pb.SignalingMessage, error) {
	b.logger.Info("Subscribe to offers from peer", zap.Any("kp", kp))

	sub, err := b.NewSubscription(kp)
	if err != nil {
		return nil, err
	}

	return sub.Channel, nil
}

func (b *Backend) Publish(kp *crypto.KeyPair, msg *pb.SignalingMessage) error {
	env, err := msg.Encrypt(kp)
	if err != nil {
		return fmt.Errorf("failed to marshal offer: %w", err)
	}

	payload, err := proto.Marshal(env)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	opts := []pubsub.PubOpt{
		pubsub.WithReadiness(pubsub.MinTopicSize(1)),
	}

	if err := b.topic.Publish(b.context, payload, opts...); err != nil {
		return fmt.Errorf("failed to publish offer: %w", err)

	}

	return nil
}

func (b *Backend) Close() error {
	b.topic.Close()

	if b.config.EnableDHTDiscovery {
		if err := b.dht.Close(); err != nil {
			return err
		}
	}

	if b.config.EnableMDNSDiscovery {
		if err := b.mdns.Close(); err != nil {
			return err
		}
	}

	if err := b.host.Close(); err != nil {
		return err
	}

	return nil
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (b *Backend) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == b.host.ID() {
		return // skip ourself
	}

	b.logger.Info("Discovered new peer via mDNS",
		zap.Any("peer", pi.ID),
		zap.Any("remotes", pi.Addrs))

	if err := b.host.Connect(b.context, pi); err != nil {
		b.logger.Error("Failed connecting to peer",
			zap.Any("peer", pi.ID),
			zap.Error(err),
		)
	}
}
