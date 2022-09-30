package wg

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/util"
	t "github.com/stv0g/cunicu/pkg/util/terminal"
	"golang.org/x/exp/slices"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Device wgtypes.Device

type DeviceList []*wgtypes.Device

func (devs *DeviceList) GetByName(name string) *wgtypes.Device {
	for _, dev := range *devs {
		if dev.Name == name {
			return dev
		}
	}

	return nil
}

func (d *Device) DumpEnv(wr io.Writer) error {
	var color, hideKeys bool

	switch os.Getenv("WG_COLOR_MODE") {
	case "always":
		color = true
	case "never":
		color = false
	case "auto":
		fallthrough
	default:
		color = t.IsATTY(os.Stdout)
	}

	if !color {
		wr = t.NewANSIStripper(wr)
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

func (d *Device) Dump(wr io.Writer, hideKeys bool) error {
	wri := t.NewIndenter(wr, "  ")

	fmt.Fprintf(wr, t.Mods("interface", t.Bold, t.FgGreen)+": "+t.Mods("%s", t.FgGreen)+"\n", d.Name)

	if crypto.Key(d.PrivateKey).IsSet() {
		t.FprintKV(wri, "public key", d.PublicKey)

		if hideKeys {
			t.FprintKV(wri, "private key", "(hidden)")
		} else {
			t.FprintKV(wri, "private key", d.PrivateKey)
		}
	}

	t.FprintKV(wri, "listening port", d.ListenPort)

	if d.FirewallMark > 0 {
		t.FprintKV(wri, "fwmark", fmt.Sprintf("%d", d.FirewallMark))
	}

	// Sort peers by last handshake time
	slices.SortFunc(d.Peers, func(a, b wgtypes.Peer) bool {
		return CmpPeerHandshakeTime(a, b) < 0
	})

	for _, p := range d.Peers {
		fmt.Fprintf(wr, "\n"+t.Mods("peer", t.Bold, t.FgYellow)+": "+t.Mods("%s", t.FgYellow)+"\n", p.PublicKey)

		if crypto.Key(p.PresharedKey).IsSet() {
			if hideKeys {
				t.FprintKV(wri, "preshared key", "(hidden)")
			} else {
				t.FprintKV(wri, "preshared key", p.PresharedKey)
			}
		}

		if p.Endpoint != nil {
			t.FprintKV(wri, "endpoint", p.Endpoint)
		}

		if !p.LastHandshakeTime.IsZero() {
			t.FprintKV(wri, "latest handshake", util.Ago(p.LastHandshakeTime))
		}

		if len(p.AllowedIPs) > 0 {
			allowedIPs := []string{}
			for _, allowedIP := range p.AllowedIPs {
				allowedIPs = append(allowedIPs, allowedIP.String())
			}

			t.FprintKV(wri, "allowed ips", strings.Join(allowedIPs, ", "))
		} else {
			t.FprintKV(wri, "allowed ips", "(none)")
		}

		if p.ReceiveBytes > 0 || p.TransmitBytes > 0 {
			t.FprintKV(wri, "transfer", fmt.Sprintf("%s received, %s sent",
				util.PrettyBytes(p.ReceiveBytes),
				util.PrettyBytes(p.TransmitBytes)))
		}

		if p.PersistentKeepaliveInterval > 0 {
			t.FprintKV(wri, "persistent keepalive", util.Every(p.PersistentKeepaliveInterval))
		}
	}

	return nil
}

func (d *Device) Config() *Config {
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
