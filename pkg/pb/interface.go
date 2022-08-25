package pb

import (
	"encoding/base64"
	"fmt"
	"io"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/pkg/util"
	t "riasc.eu/wice/pkg/util/terminal"
)

func (i *Interface) Device() *wgtypes.Device {
	peers := []wgtypes.Peer{}
	for _, peer := range i.Peers {
		peers = append(peers, peer.Peer())
	}

	return &wgtypes.Device{
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
	wri := util.NewIndenter(wr, "  ")

	if _, err := fmt.Fprintf(wr, t.Color("interface", t.Bold, t.FgGreen)+": "+t.Color("%s", t.FgGreen)+"\n", i.Name); err != nil {
		return err
	}

	t.FprintKV(wri, "public key", base64.StdEncoding.EncodeToString(i.PublicKey))
	if verbosity > 2 {
		t.FprintKV(wri, "private key", base64.StdEncoding.EncodeToString(i.PrivateKey))
	}
	t.FprintKV(wri, "listening port", i.ListenPort)

	if i.FirewallMark != 0 {
		t.FprintKV(wri, "fwmark", i.FirewallMark)
	}

	t.FprintKV(wri, "type", i.Type)
	t.FprintKV(wri, "ifindex", i.Ifindex)
	t.FprintKV(wri, "mtu", i.Mtu)
	t.FprintKV(wri, "latest sync", util.Ago(i.LastSyncTimestamp.Time()))

	if i.Ice != nil && verbosity > 3 {
		fmt.Fprintln(wr)
		i.Ice.Dump(wri, verbosity)
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
