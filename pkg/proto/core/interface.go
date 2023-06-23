// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"encoding/base64"
	"fmt"
	"io"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/stv0g/cunicu/pkg/log"
	"github.com/stv0g/cunicu/pkg/tty"
	"github.com/stv0g/cunicu/pkg/wg"
)

func (i *Interface) Device() *wg.Interface {
	peers := []wgtypes.Peer{}
	for _, peer := range i.Peers {
		peers = append(peers, peer.Peer())
	}

	return &wg.Interface{
		Name:         i.Name,
		Type:         wgtypes.DeviceType(i.Type),
		PublicKey:    *(*wgtypes.Key)(i.PublicKey),
		PrivateKey:   *(*wgtypes.Key)(i.PrivateKey),
		ListenPort:   int(i.ListenPort),
		FirewallMark: int(i.FirewallMark),
		Peers:        peers,
	}
}

// Dump writes a human readable version of the interface status to the supplied writer.
// The format resembles the one used by wg(8).
func (i *Interface) Dump(wr io.Writer, level log.Level) error {
	wri := tty.NewIndenter(wr, "  ")

	if _, err := fmt.Fprintf(wr, tty.Mods("interface", tty.Bold, tty.FgGreen)+": "+tty.Mods("%s", tty.FgGreen)+"\n", i.Name); err != nil {
		return err
	}

	if _, err := tty.FprintKV(wri, "public key", base64.StdEncoding.EncodeToString(i.PublicKey)); err != nil {
		return err
	}

	if level.Verbosity() > 5 {
		if _, err := tty.FprintKV(wri, "private key", base64.StdEncoding.EncodeToString(i.PrivateKey)); err != nil {
			return err
		}
	}

	if _, err := tty.FprintKV(wri, "listening port", i.ListenPort); err != nil {
		return err
	}

	if i.FirewallMark != 0 {
		if _, err := tty.FprintKV(wri, "fwmark", i.FirewallMark); err != nil {
			return err
		}
	}

	if _, err := tty.FprintKV(wri, "type", i.Type); err != nil {
		return err
	}

	if _, err := tty.FprintKV(wri, "ifindex", i.Ifindex); err != nil {
		return err
	}

	if _, err := tty.FprintKV(wri, "mtu", i.Mtu); err != nil {
		return err
	}

	if _, err := tty.FprintKV(wri, "latest sync", tty.Ago(i.LastSyncTimestamp.Time())); err != nil {
		return err
	}

	if i.Ice != nil && level.Verbosity() > 3 {
		if _, err := fmt.Fprintln(wr); err != nil {
			return err
		}

		if err := i.Ice.Dump(wri, level); err != nil {
			return err
		}
	}

	for _, p := range i.Peers {
		if _, err := fmt.Fprintln(wr); err != nil {
			return err
		}

		if err := p.Dump(wri, level); err != nil {
			return err
		}
	}

	return nil
}

// Redact redacts any sensitive information from the interface status such as private keys
func (i *Interface) Redact() *Interface {
	i.PrivateKey = nil

	return i
}
