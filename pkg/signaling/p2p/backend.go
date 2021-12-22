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
	log "github.com/sirupsen/logrus"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling"
	"riasc.eu/wice/pkg/socket"
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
	logger log.FieldLogger
	config BackendConfig

	peers     map[crypto.PublicKeyPair]*Peer
	peersLock sync.Mutex

	host host.Host

	context context.Context

	mdns   mdns.Service
	dht    *dht.IpfsDHT
	pubsub *pubsub.PubSub
	server *socket.Server
}

func NewBackend(uri *url.URL, server *socket.Server) (signaling.Backend, error) {
	var err error

	logFields := log.Fields{
		"logger":  "backend",
		"backend": uri.Scheme,
	}

	b := &Backend{
		peers:  map[crypto.PublicKeyPair]*Peer{},
		logger: log.WithFields(logFields),
		config: defaultConfig,
		server: server,
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
	b.logger.WithFields(log.Fields{
		"id":    b.host.ID(),
		"addrs": b.host.Addrs(),
	}).Info("Host created")

	b.host.Network().Notify(b)

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

		if b.dht, err = dht.New(b.context, b.host); err != nil {
			return nil, fmt.Errorf("failed to create DHT: %w", err)
		}

		if err = b.dht.Bootstrap(b.context); err != nil {
			return nil, fmt.Errorf("failed to bootstrap DHT: %w", err)
		}
	}

	rt := b.dht.RoutingTable()

	// Add some handlers
	rt.PeerAdded = func(i peer.ID) {
		b.logger.WithField("peer", i).Debug("Peer added to routing table")
	}

	rt.PeerRemoved = func(i peer.ID) {
		b.logger.WithField("peer", i).Debug("Peer removed from routing table")
	}

	// Let's connect to the bootstrap nodes first. They will tell us about the
	// other nodes in the network.
	var wg sync.WaitGroup
	for _, pi := range b.config.BootstrapPeers {
		b.logger.WithField("peer", pi).Debug("Connecting to peer")
		wg.Add(1)
		go func(pi peer.AddrInfo) {
			defer wg.Done()

			logger := b.logger.WithField("peer", pi)

			if err := b.host.Connect(b.context, pi); err != nil {
				logger.Warning("Failed to connect to boostrap node")
			} else {
				logger.Info("Connection established with bootstrap node")
			}
		}(pi)
	}
	wg.Wait() // TODO: can we run this asynchronously?

	rd := discovery.NewRoutingDiscovery(b.dht)

	// setup PubSub service using the GossipSub router
	if b.pubsub, err = pubsub.NewGossipSub(b.context, b.host, pubsub.WithDiscovery(rd)); err != nil {
		return nil, fmt.Errorf("failed to create pubsub router: %w", err)
	}

	as := []string{}
	for _, a := range b.host.Addrs() {
		as = append(as, a.String())
	}

	b.server.BroadcastEvent(&pb.Event{
		Type:  "backend",
		State: "ready",
		Event: &pb.Event_Backend{
			Backend: &pb.BackendEvent{
				Id:              b.host.ID().String(),
				ListenAddresses: as,
			},
		},
	})

	return b, nil
}

func (b *Backend) getPeer(kp crypto.PublicKeyPair) (*Peer, error) {
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

func (b *Backend) SubscribeOffer(kp crypto.PublicKeyPair) (chan signaling.Offer, error) {
	b.logger.WithField("kp", kp).Info("Subscribe to offers from peer")

	p, err := b.getPeer(kp)
	if err != nil {
		return nil, err
	}

	return p.Offers, nil
}

func (b *Backend) PublishOffer(kp crypto.PublicKeyPair, offer signaling.Offer) error {
	p, err := b.getPeer(kp)
	if err != nil {
		return fmt.Errorf("failed to get peer: %w", err)
	}

	return p.publishOffer(offer)
}

func (b *Backend) Close() error {
	return nil // TODO
}

func (b *Backend) Tick() {

}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (b *Backend) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == b.host.ID() {
		return // skip ourself
	}

	b.logger.WithField("peer", pi.ID).Info("Discovered new peer via mDNS")

	if err := b.host.Connect(b.context, pi); err != nil {
		b.logger.WithField("peer", pi.ID).WithError(err).Error("Failed connecting to peer")
	}
}
