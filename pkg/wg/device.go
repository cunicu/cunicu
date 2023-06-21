// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package wg

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/exp/slices"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/tty"
)

type Interface wgtypes.Device

func (d *Interface) DumpEnv(wr io.Writer) error {
	var color, hideKeys bool

	switch os.Getenv("WG_COLOR_MODE") {
	case "always":
		color = true
	case "never":
		color = false
	case "auto":
		fallthrough
	default:
		color = tty.IsATTY(os.Stdout)
	}

	if !color {
		wr = tty.NewANSIStripper(wr)
	}

	switch os.Getenv("WG_HIDE_KEYS") {
	case "never":
		hideKeys = false
	case "always":
		fallthrough
	default:
		hideKeys = true
	}

	return d.Dump(wr, hideKeys)
}

func (d *Interface) Dump(wr io.Writer, hideKeys bool) error { //nolint:gocognit
	wri := tty.NewIndenter(wr, "  ")

	fmt.Fprintf(wr, tty.Mods("interface", tty.Bold, tty.FgGreen)+": "+tty.Mods("%s", tty.FgGreen)+"\n", d.Name)

	if crypto.Key(d.PrivateKey).IsSet() {
		if _, err := tty.FprintKV(wri, "public key", d.PublicKey); err != nil {
			return err
		}

		if hideKeys {
			if _, err := tty.FprintKV(wri, "private key", "(hidden)"); err != nil {
				return err
			}
		} else {
			if _, err := tty.FprintKV(wri, "private key", d.PrivateKey); err != nil {
				return err
			}
		}
	}

	if _, err := tty.FprintKV(wri, "listening port", d.ListenPort); err != nil {
		return err
	}

	if d.FirewallMark > 0 {
		if _, err := tty.FprintKV(wri, "fwmark", fmt.Sprintf("%d", d.FirewallMark)); err != nil {
			return err
		}
	}

	// Sort peers by last handshake time
	slices.SortFunc(d.Peers, func(a, b wgtypes.Peer) bool {
		return CmpPeerHandshakeTime(a, b) < 0
	})

	for _, p := range d.Peers {
		fmt.Fprintf(wr, "\n"+tty.Mods("peer", tty.Bold, tty.FgYellow)+": "+tty.Mods("%s", tty.FgYellow)+"\n", p.PublicKey)

		if crypto.Key(p.PresharedKey).IsSet() {
			if hideKeys {
				if _, err := tty.FprintKV(wri, "preshared key", "(hidden)"); err != nil {
					return err
				}
			} else {
				if _, err := tty.FprintKV(wri, "preshared key", p.PresharedKey); err != nil {
					return err
				}
			}
		}

		if p.Endpoint != nil {
			if _, err := tty.FprintKV(wri, "endpoint", p.Endpoint); err != nil {
				return err
			}
		}

		if !p.LastHandshakeTime.IsZero() {
			if _, err := tty.FprintKV(wri, "latest handshake", tty.Ago(p.LastHandshakeTime)); err != nil {
				return err
			}
		}

		if len(p.AllowedIPs) > 0 {
			allowedIPs := []string{}
			for _, allowedIP := range p.AllowedIPs {
				allowedIPs = append(allowedIPs, allowedIP.String())
			}

			if _, err := tty.FprintKV(wri, "allowed ips", strings.Join(allowedIPs, ", ")); err != nil {
				return err
			}
		} else {
			if _, err := tty.FprintKV(wri, "allowed ips", "(none)"); err != nil {
				return err
			}
		}

		if p.ReceiveBytes > 0 || p.TransmitBytes > 0 {
			if _, err := tty.FprintKV(wri, "transfer", fmt.Sprintf("%s received, %s sent",
				tty.PrettyBytes(p.ReceiveBytes),
				tty.PrettyBytes(p.TransmitBytes))); err != nil {
				return err
			}
		}

		if p.PersistentKeepaliveInterval > 0 {
			if _, err := tty.FprintKV(wri, "persistent keepalive", tty.Every(p.PersistentKeepaliveInterval)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *Interface) Config() *Config {
	cfg := &Config{}

	if crypto.Key(d.PrivateKey).IsSet() {
		cfg.PrivateKey = &d.PrivateKey
	}

	if d.ListenPort != 0 {
		cfg.ListenPort = &d.ListenPort
	}

	if d.FirewallMark != 0 {
		cfg.FirewallMark = &d.FirewallMark
	}

	for _, p := range d.Peers {
		p := p

		pcfg := wgtypes.PeerConfig{
			PublicKey:  p.PublicKey,
			Endpoint:   p.Endpoint,
			AllowedIPs: p.AllowedIPs,
		}

		if crypto.Key(p.PresharedKey).IsSet() {
			pcfg.PresharedKey = &p.PresharedKey
		}

		if pki := p.PersistentKeepaliveInterval; pki > 0 {
			pcfg.PersistentKeepaliveInterval = &pki
		}

		cfg.Peers = append(cfg.Peers, pcfg)
	}

	return cfg
}
