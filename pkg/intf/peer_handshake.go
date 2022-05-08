package intf

import (
	"fmt"
	"net"
	"time"
)

// ensureHandshake initiated a new Wireguard handshake if the last one is older than 5 seconds
func (p *Peer) ensureHandshake() error {
	// Return if the last handshake happed within the last 5 seconds
	if time.Since(p.LastHandshakeTime) < 5*time.Second {
		return nil
	}

	if err := p.initiateHandshake(); err != nil {
		return fmt.Errorf("failed to initiate handshake: %w", err)
	}

	return nil
}

// initiateHandshake sends a single packet towards the peer
// which triggers Wireguard to initiate the handshake
func (p *Peer) initiateHandshake() error {
	for time.Since(p.LastHandshakeTime) > 5*time.Second {
		p.logger.Debug("Waiting for handshake")

		ra := &net.UDPAddr{
			IP:   p.PublicKey().IPv6Address().IP,
			Zone: p.Interface.Name(),
			Port: 1234,
		}

		c, err := net.DialUDP("udp6", nil, ra)
		if err != nil {
			return err
		}

		if _, err := c.Write([]byte{1}); err != nil {
			return err
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}
