package p2p

import (
	"fmt"
	"strconv"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/pnet"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	maddr "github.com/multiformats/go-multiaddr"
	"riasc.eu/wice/pkg/signaling"
)

var defaultConfig = BackendConfig{
	Private:             false,
	EnableDHTDiscovery:  true,
	EnableMDNSDiscovery: true,
	ListenAddresses:     make(multiAddressList, 0),
	BootstrapPeers:      make(peerAddressList, 0),
}

type peerAddressList []peer.AddrInfo
type multiAddressList []maddr.Multiaddr

type BackendConfig struct {
	signaling.BackendConfig

	// Load some options
	ListenAddresses multiAddressList

	// BootstrapPeers is a list of peers to which we initially connect
	BootstrapPeers peerAddressList

	// PrivateKey is the private key used by the libp2p host.
	PrivateKey crypto.PrivKey

	// PrivateNetwork configures libp2p to use the given private network protector.
	PrivateNetwork pnet.PSK

	// DHTDiscovery enables peer discovery and content routing via the Kadmelia DHT.
	EnableDHTDiscovery bool

	// MDNSDiscovery enables peer discovery via local mDNS.
	EnableMDNSDiscovery bool

	// NATPortMap configures libp2p to use the default NATManager. The default NATManager will attempt to open a port in your network's firewall using UPnP.
	EnableNATPortMap bool

	// Do not connect to public bootstrap peers
	Private bool
}

func (al *multiAddressList) Set(as []string) error {
	for _, a := range as {
		ma, err := maddr.NewMultiaddr(a)
		if err != nil {
			return err
		}

		*al = append(*al, ma)
	}

	return nil
}

func (al *peerAddressList) Set(as []string) error {
	for _, a := range as {
		pi, err := peer.AddrInfoFromString(a)
		if err != nil {
			return err
		}

		*al = append(*al, *pi)
	}

	return nil
}

func (c *BackendConfig) Parse(cfg *signaling.BackendConfig) error {
	var err error

	options := cfg.URI.Query()

	c.BackendConfig = *cfg

	if pkStr := options.Get("private-key"); pkStr != "" {
		pk, err := crypto.ConfigDecodeKey(pkStr)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}

		if c.PrivateKey, err = crypto.UnmarshalEd25519PrivateKey(pk); err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	if bStr := options.Get("mdns"); bStr != "" {
		if c.EnableMDNSDiscovery, err = strconv.ParseBool(bStr); err != nil {
			return fmt.Errorf("failed to parse mdns option: %w", err)
		}
	}

	if bStr := options.Get("dht"); bStr != "" {
		if c.EnableDHTDiscovery, err = strconv.ParseBool(bStr); err != nil {
			return fmt.Errorf("failed to parse dht option: %w", err)
		}
	}

	if laStrs, ok := options["listen-address"]; ok {
		if err := c.ListenAddresses.Set(laStrs); err != nil {
			return fmt.Errorf("failed to parse listen-address option: %w", err)
		}
	}

	if bpStrs, ok := options["bootstrap-peer"]; ok {
		if err := c.BootstrapPeers.Set(bpStrs); err != nil {
			return fmt.Errorf("failed to parse listen-address option: %w", err)
		}
	}

	if privateStrs, ok := options["private"]; ok {
		if c.Private, err = strconv.ParseBool(privateStrs[0]); err != nil {
			return fmt.Errorf("failed to parse %s as a boolean value: %w", privateStrs[0], err)
		}
	}

	// use the default set of bootstrap peers if none are provided
	if len(c.BootstrapPeers) == 0 {
		for _, s := range dht.DefaultBootstrapPeers {
			if pi, err := peer.AddrInfoFromP2pAddr(s); err == nil {
				c.BootstrapPeers = append(c.BootstrapPeers, *pi)
			}
		}
	}

	return nil
}
