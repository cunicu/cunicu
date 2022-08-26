package wg

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/exp/slices"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/pkg/util"
	t "riasc.eu/wice/pkg/util/terminal"
)

type Device wgtypes.Device

type Devices []*wgtypes.Device

func (devs *Devices) GetByName(name string) *wgtypes.Device {
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
		color = util.IsATTY()
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

	if _, err := fmt.Fprintf(wr, t.Color("interface", t.Bold, t.FgGreen)+": "+t.Color("%s", t.FgGreen)+"\n", d.Name); err != nil {
		return err
	}

	t.FprintKV(wri, "public key", d.PublicKey)
	t.FprintKV(wri, "private key", "(hidden)")
	t.FprintKV(wri, "listening port", d.ListenPort)

	if !hideKeys {
		t.FprintKV(wri, "private key", d.PrivateKey)
	}

	if d.FirewallMark > 0 {
		t.FprintKV(wri, "fwmark", fmt.Sprintf("%#x", d.FirewallMark))
	}

	// Sort peers by last handshake time
	slices.SortFunc(d.Peers, func(a, b wgtypes.Peer) bool { return CmpPeerHandshakeTime(&a, &b) < 0 })

	for _, p := range d.Peers {
		if _, err := fmt.Fprintf(wr, " \n"+t.Color("peer", t.Bold, t.FgYellow)+": "+t.Color("%s", t.FgYellow)+"\n", p.PublicKey.String()); err != nil {
			return err
		}

		if p.Endpoint != nil {
			t.FprintKV(wri, "endpoint", p.Endpoint)
		}

		if p.LastHandshakeTime.Second() > 0 {
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
	zero := wgtypes.Key{}

	cfg := &Config{}

	if d.PrivateKey != zero {
		cfg.PrivateKey = &d.PrivateKey
	}

	if d.ListenPort != 0 {
		cfg.ListenPort = &d.ListenPort
	}

	if d.FirewallMark != 0 {
		cfg.FirewallMark = &d.FirewallMark
	}

	for _, p := range d.Peers {
		pcfg := wgtypes.PeerConfig{
			PublicKey:  p.PublicKey,
			Endpoint:   p.Endpoint,
			AllowedIPs: p.AllowedIPs,
		}

		if p.PresharedKey != zero {
			pcfg.PresharedKey = &p.PresharedKey
		}

		if pki := p.PersistentKeepaliveInterval; pki > 0 {
			pcfg.PersistentKeepaliveInterval = &pki
		}

		cfg.Peers = append(cfg.Peers, pcfg)
	}

	return cfg
}
