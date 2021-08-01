package p2p

import (
	"fmt"
	"net/url"
	"strings"

	"riasc.eu/wice/pkg/backend/base"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	maddr "github.com/multiformats/go-multiaddr"
)

type addressList []maddr.Multiaddr

type BackendConfig struct {
	base.BackendConfig

	// Load some options
	ListenAddresses  addressList
	BootstrapPeers   addressList
	RendezvousString string
}

func (al addressList) Set(option string) error {
	as := strings.Split(option, ":")
	for _, a := range as {
		ma, err := maddr.NewMultiaddr(a)
		if err != nil {
			return err
		}

		al = append(al, ma)
	}

	return nil
}

func (c *BackendConfig) Parse(uri *url.URL, options map[string]string) error {
	if rStr, ok := options["rendevouz-string"]; ok {
		c.RendezvousString = rStr
	} else {
		c.RendezvousString = uri.Host
	}

	if laStr, ok := options["listen-addresses"]; ok {
		err := c.ListenAddresses.Set(laStr)
		if err != nil {
			return fmt.Errorf("failed to parse listen-address option: %w", err)
		}
	}

	if bpStr, ok := options["bootstrap-peers"]; ok {
		if err := c.BootstrapPeers.Set(bpStr); err != nil {
			return fmt.Errorf("failed to parse listen-address option: %w", err)
		}
	}

	if len(c.BootstrapPeers) == 0 {
		c.BootstrapPeers = dht.DefaultBootstrapPeers
	}

	return nil
}
