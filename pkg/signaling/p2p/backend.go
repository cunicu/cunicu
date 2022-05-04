package p2p

import (
	"context"
	"fmt"
	"net"
	"sync"

	p2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	ma "github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
	"google.golang.org/protobuf/reflect/protoreflect"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/pb/uvarint"
	"riasc.eu/wice/pkg/signaling"
)

const (
	userAgent = "wice"

	protocolSignaling      = "/wice/signaling/0.1.0"
	protocolIdentification = "/wice/id/0.1.0"
)

func init() {
	signaling.Backends["p2p"] = &signaling.BackendPlugin{
		New:         NewBackend,
		Description: "libp2p",
	}
}

type Backend struct {
	signaling.SubscriptionsRegistry

	peers     map[crypto.Key][]peer.ID
	peersLock sync.RWMutex
	peersCond *sync.Cond

	logger *zap.Logger
	config BackendConfig

	context context.Context

	host host.Host
	mdns mdns.Service
	dht  *dht.IpfsDHT

	events chan *pb.Event
}

func NewBackend(cfg *signaling.BackendConfig, events chan *pb.Event, logger *zap.Logger) (signaling.Backend, error) {
	var err error

	b := &Backend{
		SubscriptionsRegistry: signaling.NewSubscriptionsRegistry(),

		peers: map[crypto.Key][]peer.ID{},

		logger:  logger,
		config:  defaultConfig,
		events:  events,
		context: context.Background(),
	}

	b.peersCond = sync.NewCond(&b.peersLock)

	if err := b.config.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse backend options: %w", err)
	}

	opts := b.options()
	b.host, err = p2p.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	b.logger.Info("Host created",
		zap.Any("id", b.host.ID()),
		zap.Any("addrs", b.StringAddrs()),
	)

	b.host.SetStreamHandler(protocolIdentification, b.handleIdentificationStream)
	b.host.SetStreamHandler(protocolSignaling, b.handleSignalingStream)

	if b.config.EnableMDNSDiscovery {
		b.logger.Debug("Setup mDNS discovery")

		b.mdns = mdns.NewMdnsService(b.host, mdnsServiceName, &mDNSNotifee{b})
		if err := b.mdns.Start(); err != nil {
			return nil, fmt.Errorf("failed to start mDNS service: %w", err)
		}
	}

	if b.config.EnableDHTDiscovery && b.dht != nil {
		b.logger.Debug("Bootstrapping the DHT")

		if err = b.dht.Bootstrap(b.context); err != nil {
			return nil, fmt.Errorf("failed to bootstrap DHT: %w", err)
		}
	}

	if !b.config.Private {
		b.bootstrap()
	}

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

func (b *Backend) Publish(ctx context.Context, kp *crypto.KeyPair, msg *pb.SignalingMessage) error {
	if err := b.waitForPeer(ctx, &kp.Theirs); err != nil {
		return fmt.Errorf("failed to wait for peer: %w", err)
	}

	env, err := msg.Encrypt(kp)
	if err != nil {
		return fmt.Errorf("failed to encrypt message: %w", err)
	}

	b.peersLock.RLock()
	if pids, ok := b.peers[kp.Theirs]; ok {
		if err := b.sendMessageToPeers(ctx, pids, env); err != nil {
			b.peersLock.RUnlock()
			return fmt.Errorf("failed to send: %w", err)
		}
	}
	b.peersLock.RUnlock()

	b.logger.Info("Published to message to peers", zap.Any("kp", kp))

	return nil
}

func (b *Backend) Subscribe(ctx context.Context, kp *crypto.KeyPair) (chan *pb.SignalingMessage, error) {
	go b.watchPeer(&kp.Theirs)

	sub, err := b.NewSubscription(kp)
	if err != nil {
		return nil, err
	}

	b.logger.Info("Subscribed to messages from peer", zap.Any("kp", kp))

	return sub.C, nil
}

func (b *Backend) Close() error {
	if b.dht != nil {
		if err := b.dht.Close(); err != nil {
			return err
		}
	}

	if b.mdns != nil {
		if err := b.mdns.Close(); err != nil {
			return err
		}
	}

	return b.host.Close()
}

func (b *Backend) handleMDNSPeer(ai peer.AddrInfo) {
	if ai.ID == b.host.ID() {
		return
	}

	if err := b.host.Connect(b.context, ai); err != nil {
		b.logger.Error("Failed to connect to mDNS peer", zap.Error(err))
		return
	}

	b.logger.Info("Found new peer via mDNS", zap.Any("peer", ai))

	b.mDNSPeersLock.Lock()
	b.mDNSPeers = append(b.mDNSPeers, ai.ID)
	b.mDNSPeersLock.Unlock()
}

func (b *Backend) options() []p2p.Option {
	opts := []p2p.Option{
		p2p.Defaults,
		p2p.UserAgent(userAgent),
		p2p.EnableNATService(),
		p2p.EnableRelay(),
		p2p.EnableRelayService(),
	}

	if len(b.config.ListenAddresses) > 0 {
		opts = append(opts, p2p.ListenAddrs([]ma.Multiaddr(b.config.ListenAddresses)...))
	} else {
		opts = append(opts, p2p.DefaultListenAddrs)
	}

	if b.config.EnableNATPortMap {
		opts = append(opts, p2p.NATPortMap())
	}

	if b.config.PrivateKey != nil {
		opts = append(opts, p2p.Identity(b.config.PrivateKey))
	}

	if b.config.PrivateNetwork != nil {
		opts = append(opts, p2p.PrivateNetwork(b.config.PrivateNetwork))
	}

	if b.config.EnableDHTDiscovery {
		opts = append(opts, p2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			var err error
			b.dht, err = dht.New(b.context, h)
			return b.dht, err
		}))
	}

	return opts
}

func (b *Backend) handleSignalingStream(s network.Stream) {
	b.logger.Info("Handle new stream", zap.Any("protocol", s.Protocol()))

	rd := uvarint.NewDelimitedReader(s, 10<<10)

	for {
		var env pb.SignalingEnvelope
		if err := rd.ReadMsg(&env); err != nil {
			b.logger.Warn("Failed to read message", zap.Error(err))
		}

		kp, err := env.PublicKeyPair()
		if err != nil {
			b.logger.Error("Failed open envelope", zap.Error(err))
			return
		}

		sub, err := b.GetSubscription(&kp)
		if err != nil {
			b.logger.Error("Failed to find matching subscription", zap.Error(err))
			return
		}

		if err := sub.NewMessage(&env); err != nil {
			b.logger.Error("Failed to handle new message", zap.Error(err))
			return
		}

		if err := s.Close(); err != nil {
			b.logger.Error("Failed to close stream", zap.Error(err))
		}
	}
}

func (b *Backend) handleIdentificationStream(s network.Stream) {
	b.GetSubscription()
}

func (b *Backend) sendMessageToPeers(ctx context.Context, pids []peer.ID, msg protoreflect.ProtoMessage) error {
	for _, pid := range pids {
		s, err := b.host.NewStream(ctx, pid, protocolSignaling)
		if err != nil {
			return err
		}

		wr := uvarint.NewDelimitedWriter(s)
		if err := wr.WriteMsg(msg); err != nil {
			return err
		}
	}

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

	b.logger.Debug("Bootstrap finished")
}

func (b *Backend) waitForPeer(ctx context.Context, pk *crypto.Key) {
	b.peersLock.RLock()
	defer b.peersLock.RUnlock()

	for {
		if pids, ok := b.peers[*pk]; ok && len(pids) > 0 {
			break
		}

		b.peersCond.Wait()
	}
}

func (b *Backend) watchPeer(pk *crypto.Key) {
	c := publicKeyToCid(pk)
	for ai := range b.dht.FindProvidersAsync(b.context, c, 0) {
		if ai.ID == b.host.ID() {
			continue // found ourself?
		}

		if ok, err := b.checkPeer(ai); err != nil {
			b.logger.Error("Failed to validate peer", zap.Error(err))
		} else if ok {
			b.addPeer(pk, ai.ID)
		}
	}
}

func (b *Backend) checkPeer(pk *crypto.Key, ai peer.AddrInfo) (bool, error) {
	if err := b.host.Connect(context.TODO(), ai); err != nil {
		b.logger.Error("Failed to connect to peeer", zap.Any("addr", ai))
	}

	s, err := b.host.NewStream(context.TODO(), ai.ID, protocolID)

	b.logger.Info("Found peer", zap.Any("addr", ai), zap.Any("peer", pk))

	return true, nil
}

func (b *Backend) addPeer(pk *crypto.Key, pid peer.ID) {
	b.peersLock.Lock()
	defer b.peersLock.Unlock()

	if ids, ok := b.peers[*pk]; ok {
		ids = append(ids, pid)
	} else {
		b.peers[*pk] = []peer.ID{pid}
	}

	b.peersCond.Broadcast()
}
