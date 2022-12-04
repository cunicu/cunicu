package core

import (
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/stv0g/cunicu/pkg/util"
	t "github.com/stv0g/cunicu/pkg/util/terminal"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func (p *Peer) Peer() wgtypes.Peer {
	allowedIPs := []net.IPNet{}
	for _, allowedIP := range p.AllowedIps {
		_, ipnet, err := net.ParseCIDR(allowedIP)
		if err != nil {
			panic(fmt.Errorf("failed to parse WireGuard AllowedIP: %w", err))
		}

		allowedIPs = append(allowedIPs, *ipnet)
	}

	endpoint, err := net.ResolveUDPAddr("udp", p.Endpoint)
	if err != nil {
		panic(fmt.Errorf("failed to parse WireGuard Endpoint: %w", err))
	}

	q := wgtypes.Peer{
		PublicKey:                   *(*wgtypes.Key)(p.PublicKey),
		Endpoint:                    endpoint,
		PersistentKeepaliveInterval: time.Duration(p.PersistentKeepaliveInterval) * time.Second,
		TransmitBytes:               p.TransmitBytes,
		ReceiveBytes:                p.ReceiveBytes,
		AllowedIPs:                  allowedIPs,
		ProtocolVersion:             int(p.ProtocolVersion),
	}

	if p.PresharedKey != nil {
		q.PresharedKey = *(*wgtypes.Key)(p.PresharedKey)
	}

	if p.LastHandshakeTimestamp != nil {
		q.LastHandshakeTime = p.LastHandshakeTimestamp.Time()
	}

	return q
}

func (p *Peer) Dump(wr io.Writer, verbosity int) error { //nolint:gocognit
	wri := t.NewIndenter(wr, "  ")

	if _, err := fmt.Fprintf(wr, t.Mods("peer", t.Bold, t.FgYellow)+": "+t.Mods("%s", t.FgYellow)+"\n", p.Name); err != nil {
		return err
	}

	if _, err := t.FprintKV(wri, "public key", base64.StdEncoding.EncodeToString(p.PublicKey)); err != nil {
		return err
	}

	if p.Endpoint != "" {
		if _, err := t.FprintKV(wri, "endpoint", p.Endpoint); err != nil {
			return err
		}
	}

	if p.Reachability != ReachabilityType_REACHABILITY_TYPE_UNSPECIFIED {
		if _, err := t.FprintKV(wri, "reachability", p.Reachability); err != nil {
			return err
		}
	}

	if p.LastHandshakeTimestamp != nil {
		if _, err := t.FprintKV(wri, "latest handshake", util.Ago(p.LastHandshakeTimestamp.Time())); err != nil {
			return err
		}
	}

	if p.LastReceiveTimestamp != nil {
		if _, err := t.FprintKV(wri, "latest receive", util.Ago(p.LastReceiveTimestamp.Time())); err != nil {
			return err
		}
	}

	if p.LastTransmitTimestamp != nil {
		if _, err := t.FprintKV(wri, "latest transmit", util.Ago(p.LastTransmitTimestamp.Time())); err != nil {
			return err
		}
	}

	if len(p.AllowedIps) > 0 {
		if _, err := t.FprintKV(wri, "allowed ips", strings.Join(p.AllowedIps, ", ")); err != nil {
			return err
		}
	} else {
		if _, err := t.FprintKV(wri, "allowed ips", "(none)"); err != nil {
			return err
		}
	}

	if p.ReceiveBytes > 0 || p.TransmitBytes > 0 {
		if _, err := t.FprintKV(wri, "transfer", fmt.Sprintf("%s received, %s sent",
			util.PrettyBytes(p.ReceiveBytes),
			util.PrettyBytes(p.TransmitBytes))); err != nil {
			return err
		}
	}

	if p.PersistentKeepaliveInterval > 0 {
		if _, err := t.FprintKV(wri, "persistent keepalive", util.Every(time.Duration(p.PersistentKeepaliveInterval)*time.Second)); err != nil {
			return err
		}
	}

	if len(p.PresharedKey) > 0 && verbosity > 5 {
		if _, err := t.FprintKV(wri, "preshared key", base64.StdEncoding.EncodeToString(p.PresharedKey)); err != nil {
			return err
		}
	}

	if _, err := t.FprintKV(wri, "protocol version", p.ProtocolVersion); err != nil {
		return err
	}

	if p.Ice != nil {
		if _, err := fmt.Fprintln(wr); err != nil {
			return err
		}

		if err := p.Ice.Dump(wri, verbosity); err != nil {
			return err
		}
	}

	return nil
}

// Redact redacts any sensitive information from the peer status such as the preshared key
func (p *Peer) Redact() *Peer {
	p.PresharedKey = nil

	return p
}
