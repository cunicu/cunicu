package pb

import (
	"fmt"
	"net"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func NewPeer(p wgtypes.Peer) *Peer {
	allowedIPs := []string{}
	for _, allowedIP := range p.AllowedIPs {
		allowedIPs = append(allowedIPs, allowedIP.String())
	}

	q := &Peer{
		PublicKey:                   p.PublicKey[:],
		Endpoint:                    p.Endpoint.String(),
		PresharedKey:                p.PresharedKey[:],
		PersistentKeepaliveInterval: uint32(p.PersistentKeepaliveInterval),
		TransmitBytes:               p.TransmitBytes,
		ReceiveBytes:                p.ReceiveBytes,
		AllowedIps:                  allowedIPs,
		ProtocolVersion:             uint32(p.ProtocolVersion),
	}

	if !p.LastHandshakeTime.IsZero() {
		q.LastHandshake = Time(p.LastHandshakeTime)
	}

	return q
}

func (p *Peer) Peer() wgtypes.Peer {
	allowedIPs := []net.IPNet{}
	for _, allowedIP := range p.AllowedIps {
		_, ipnet, err := net.ParseCIDR(allowedIP)
		if err != nil {
			panic(fmt.Errorf("failed to parse Wireguard AllowedIP: %w", err))
		}

		allowedIPs = append(allowedIPs, *ipnet)
	}

	endpoint, err := net.ResolveUDPAddr("udp", p.Endpoint)
	if err != nil {
		panic(fmt.Errorf("failed to parse Wireguard Endpoint: %w", err))
	}

	q := wgtypes.Peer{
		PublicKey:                   *(*wgtypes.Key)(p.PublicKey),
		PresharedKey:                *(*wgtypes.Key)(p.PresharedKey),
		Endpoint:                    endpoint,
		PersistentKeepaliveInterval: time.Duration(p.PersistentKeepaliveInterval * uint32(time.Second)),
		TransmitBytes:               p.TransmitBytes,
		ReceiveBytes:                p.ReceiveBytes,
		AllowedIPs:                  allowedIPs,
		ProtocolVersion:             int(p.ProtocolVersion),
	}

	if p.LastHandshake != nil {
		q.LastHandshakeTime = p.LastHandshake.Time()
	}

	return q
}
