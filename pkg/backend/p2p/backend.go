package p2p

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"github.com/ipfs/go-cid"
	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
	"riasc.eu/wice/pkg/backend"
	"riasc.eu/wice/pkg/backend/base"
	"riasc.eu/wice/pkg/crypto"

	log "github.com/sirupsen/logrus"
)

const (
	ProtocolID = "/wice/rpc/0.1.0"

	// See: https://github.com/multiformats/multicodec/blob/master/table.csv#L85
	CodeX25519PublicKey = 0xec
)

func init() {
	backend.Backends["p2p"] = &backend.BackendPlugin{
		New:         NewBackend,
		Description: "LibP2P Kademlia DHT",
	}
}

type Backend struct {
	base.Backend
	config BackendConfig

	host  host.Host
	peers PeerList

	context context.Context
	dht     *dht.IpfsDHT
}

func NewBackend(uri *url.URL, options map[string]string) (backend.Backend, error) {
	var err error
	b := Backend{
		Backend: base.NewBackend(uri, options),
	}

	b.config.Parse(uri, options)

	b.context = context.Background()

	b.host, err = libp2p.New(b.context, libp2p.ListenAddrs([]multiaddr.Multiaddr(b.config.ListenAddresses)...))
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}
	b.Logger.WithField("id", b.host.ID()).WithField("addrs", b.host.Addrs()).Info("Host created")

	b.host.SetStreamHandler(ProtocolID, b.handleStream)

	b.dht, err = dht.New(b.context, b.host)
	if err != nil {
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}

	b.Logger.Debug("Bootstrapping the DHT")
	if err = b.dht.Bootstrap(b.context); err != nil {
		return nil, fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	// Let's connect to the bootstrap nodes first. They will tell us about the
	// other nodes in the network.
	var wg sync.WaitGroup
	for _, peerAddr := range b.config.BootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := b.host.Connect(b.context, *peerinfo); err != nil {
				log.Warning(err)
			} else {
				b.Logger.WithField("peer", *peerinfo).Info("Connection established with bootstrap node")
			}
		}()
	}
	wg.Wait()

	// // We use a rendezvous point "meet me here" to announce our location.
	// // This is like telling your friends to meet you at the Eiffel Tower.
	// b.Logger.Info("Announcing ourselves...")
	// routingDiscovery := discovery.NewRoutingDiscovery(b.dht)

	// discovery.Advertise(b.context, routingDiscovery, b.config.RendezvousString)
	// b.Logger.Debug("Successfully announced!")

	// // Now, look for others who have announced
	// // This is like your friend telling you the location to meet you.
	// b.Logger.Debug("Searching for other peers...")
	// peerChan, err := routingDiscovery.FindPeers(b.context, b.config.RendezvousString)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to find peers: %w", err)
	// }

	// go b.handlePeers(peerChan)

	return &b, nil
}

func (b *Backend) SubscribeOffer(kp crypto.PublicKeyPair) (chan backend.Offer, error) {
	ch := b.Backend.SubscribeOffers(kp)

	cid := cidForPublicKey(kp.Ours)
	err := b.dht.Provide(b.context, cid, true)
	if err != nil {
		return nil, err
	}

	return ch, nil
}

func (b *Backend) PublishOffer(kp crypto.PublicKeyPair, offer backend.Offer) error {
	cid := cidForPublicKey(kp.Theirs)

	peerChan := b.dht.FindProvidersAsync(b.context, cid, 0)
	go func() {
		for pai := range peerChan {
			peer, err := NewPeer(b, pai.ID)
			if err != nil {
				b.Logger.WithError(err).Error("Failed to create peer")
				return
			}

			om := backend.OfferMap{
				kp.Ours: offer,
			}

			var ret bool
			err = peer.Client.Call("candidates.Publish", &om, &ret)
			if err != nil {
				b.Logger.WithError(err).Error("Failed RPC call")
			}

			b.peers = append(b.peers, peer)
		}
	}()

	return b.PublishOffer(kp, offer)
}

func (b *Backend) Close() error {
	return nil // TODO
}

func cidForPublicKey(pk crypto.Key) cid.Cid {
	mh, _ := multihash.Encode(pk[:], multihash.IDENTITY)

	return cid.NewCidV1(CodeX25519PublicKey, mh)
}
