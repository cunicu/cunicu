package e2e

import (
	"fmt"
	"math/rand"
	"net"
	"net/url"

	"github.com/libp2p/go-libp2p-core/peer"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
)

func GenerateKeyPairs() (*crypto.KeyPair, *crypto.KeyPair, error) {
	ourKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		return nil, nil, err
	}

	theirKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		return nil, nil, err
	}

	return &crypto.KeyPair{
			Ours:   ourKey,
			Theirs: theirKey.PublicKey(),
		}, &crypto.KeyPair{
			Ours:   theirKey,
			Theirs: ourKey.PublicKey(),
		}, nil
}

func GenerateSignalingMessage() *pb.SignalingMessage {
	return &pb.SignalingMessage{
		Msg: &pb.SignalingMessage_Offer{
			Offer: &pb.SessionDescription{
				Epoch: rand.Int63(),
			},
		},
	}
}

func ParseIP(s string) (net.IPNet, error) {
	ip, netw, err := net.ParseCIDR(s)
	if err != nil {
		return net.IPNet{}, err
	}

	return net.IPNet{
		IP:   ip,
		Mask: netw.Mask,
	}, nil
}

func SignalingURL(nodes []*Agent) (*url.URL, error) {
	q := url.Values{}
	// q.Add("dht", "false")
	q.Add("mdns", "false")

	for _, node := range nodes {
		pi := &peer.AddrInfo{
			ID:    node.ID,
			Addrs: node.ListenAddresses,
		}

		mas, err := peer.AddrInfoToP2pAddrs(pi)
		if err != nil {
			return nil, fmt.Errorf("failed to get p2p addresses")
		}

		for _, ma := range mas {
			q.Add("bootstrap-peer", ma.String())
		}
	}

	return &url.URL{
		Scheme:   "p2p",
		RawQuery: q.Encode(),
	}, nil
}
