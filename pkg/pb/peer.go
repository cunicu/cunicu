package pb

import (
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/pkg/util"
	t "riasc.eu/wice/pkg/util/terminal"
)

func NewInterfaceType(t wgtypes.DeviceType) PeerDescription_InterfaceType {
	switch t {
	case wgtypes.LinuxKernel:
		return PeerDescription_LINUX_KERNEL
	case wgtypes.OpenBSDKernel:
		return PeerDescription_OPENBSD_KERNEL
	case wgtypes.WindowsKernel:
		return PeerDescription_WINDOWS_KERNEL
	case wgtypes.Userspace:
		return PeerDescription_USERSPACE
	}

	return PeerDescription_UNKNOWN
}

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
		PresharedKey:                *(*wgtypes.Key)(p.PresharedKey),
		Endpoint:                    endpoint,
		PersistentKeepaliveInterval: time.Duration(p.PersistentKeepaliveInterval * uint32(time.Second)),
		TransmitBytes:               p.TransmitBytes,
		ReceiveBytes:                p.ReceiveBytes,
		AllowedIPs:                  allowedIPs,
		ProtocolVersion:             int(p.ProtocolVersion),
	}

	if p.LastHandshakeTimestamp != nil {
		q.LastHandshakeTime = p.LastHandshakeTimestamp.Time()
	}

	return q
}

func (p *Peer) Dump(wr io.Writer, verbosity int) error {
	wri := util.NewIndenter(wr, "  ")

	if _, err := fmt.Fprintf(wr, t.Color("peer", t.Bold, t.FgYellow)+": "+t.Color("%s", t.FgYellow)+"\n", base64.StdEncoding.EncodeToString(p.PublicKey)); err != nil {
		return err
	}

	if p.Name != "" {
		if _, err := t.FprintKV(wri, "name", p.Name); err != nil {
			return err
		}
	}

	if p.Endpoint != "" {
		if _, err := t.FprintKV(wri, "endpoint", p.Endpoint); err != nil {
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

	if len(p.PresharedKey) > 0 {
		if _, err := t.FprintKV(wri, "preshared key", base64.StdEncoding.EncodeToString(p.PresharedKey)); err != nil {
			return err
		}
	}

	if _, err := t.FprintKV(wri, "protocol version", p.ProtocolVersion); err != nil {
		return err
	}

	if p.Ice != nil && verbosity > 4 {
		if _, err := fmt.Fprintln(wr); err != nil {
			return err
		}

		if err := p.Ice.Dump(wri, verbosity); err != nil {
			return err
		}
	}

	return nil
}
