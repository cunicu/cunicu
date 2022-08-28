package core

import (
	"encoding/base64"
	"fmt"
	"io"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/pkg/util"
	t "riasc.eu/wice/pkg/util/terminal"
	"riasc.eu/wice/pkg/wg"
)

func (i *Interface) Device() *wg.Device {
	peers := []wgtypes.Peer{}
	for _, peer := range i.Peers {
		peers = append(peers, peer.Peer())
	}

	return &wg.Device{
		Name:         i.Name,
		Type:         wgtypes.DeviceType(i.Type),
		PublicKey:    *(*wgtypes.Key)(i.PublicKey),
		PrivateKey:   *(*wgtypes.Key)(i.PrivateKey),
		ListenPort:   int(i.ListenPort),
		FirewallMark: int(i.FirewallMark),
		Peers:        peers,
	}
}

func (i *Interface) Dump(wr io.Writer, verbosity int) error {
	wri := t.NewIndenter(wr, "  ")

	if _, err := fmt.Fprintf(wr, t.Color("interface", t.Bold, t.FgGreen)+": "+t.Color("%s", t.FgGreen)+"\n", i.Name); err != nil {
		return err
	}

	if _, err := t.FprintKV(wri, "public key", base64.StdEncoding.EncodeToString(i.PublicKey)); err != nil {
		return err
	}

	if verbosity > 2 {
		if _, err := t.FprintKV(wri, "private key", base64.StdEncoding.EncodeToString(i.PrivateKey)); err != nil {
			return err
		}
	}

	if _, err := t.FprintKV(wri, "listening port", i.ListenPort); err != nil {
		return err
	}

	if i.FirewallMark != 0 {
		if _, err := t.FprintKV(wri, "fwmark", i.FirewallMark); err != nil {
			return err
		}
	}

	if _, err := t.FprintKV(wri, "type", i.Type); err != nil {
		return err
	}

	if _, err := t.FprintKV(wri, "ifindex", i.Ifindex); err != nil {
		return err
	}

	if _, err := t.FprintKV(wri, "mtu", i.Mtu); err != nil {
		return err
	}

	if _, err := t.FprintKV(wri, "latest sync", util.Ago(i.LastSyncTimestamp.Time())); err != nil {
		return err
	}

	if i.Ice != nil && verbosity > 3 {
		if _, err := fmt.Fprintln(wr); err != nil {
			return err
		}

		if err := i.Ice.Dump(wri, verbosity); err != nil {
			return err
		}
	}

	for _, p := range i.Peers {
		if _, err := fmt.Fprintln(wr); err != nil {
			return err
		}

		if err := p.Dump(wri, verbosity); err != nil {
			return err
		}
	}

	return nil
}
