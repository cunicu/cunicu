package wg

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/internal/util"
	t "riasc.eu/wice/internal/util/terminal"
)

type Device wgtypes.Device

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

	switch os.Getenv("WG_HIDE_KEYS") {
	case "never":
		hideKeys = false
	case "always":
		fallthrough
	default:
		hideKeys = true
	}

	return d.Dump(wr, color, hideKeys)
}

func (d *Device) Dump(wr io.Writer, color bool, hideKeys bool) error {
	var kv = map[string]any{
		"public key":     d.PublicKey,
		"private key":    "(hidden)",
		"listening port": d.ListenPort,
	}

	if !hideKeys {
		kv["private key"] = d.PrivateKey
	}

	if d.FirewallMark > 0 {
		kv["fwmark"] = fmt.Sprintf("%#x", d.FirewallMark)
	}

	if _, err := t.FprintfColored(wr, color, t.Color("interface", t.Bold, t.FgGreen)+": "+t.Color("%s", t.FgGreen)+"\n", d.Name); err != nil {
		return err
	}
	if _, err := t.PrintKeyValues(wr, color, "  ", kv); err != nil {
		return err
	}

	// TODO: sort peer list
	// https://github.com/WireGuard/wireguard-tools/blob/1fd95708391088742c139010cc6b821add941dec/src/show.c#L47

	for _, peer := range d.Peers {
		var kv = map[string]any{
			"allowed ips": "(none)",
		}

		if peer.Endpoint != nil {
			kv["endpoint"] = peer.Endpoint
		}

		if peer.LastHandshakeTime.Second() > 0 {
			kv["latest handshake"] = util.Ago(peer.LastHandshakeTime, color)
		}

		if len(peer.AllowedIPs) > 0 {
			allowedIPs := []string{}
			for _, allowedIP := range peer.AllowedIPs {
				allowedIPs = append(allowedIPs, allowedIP.String())
			}

			kv["allowed ips"] = strings.Join(allowedIPs, ", ")
		} else {
			kv["allowed ips"] = "(none)"
		}

		if peer.ReceiveBytes > 0 || peer.TransmitBytes > 0 {
			kv["transfer"] = fmt.Sprintf("%s received, %s sent\n",
				util.PrettyBytes(peer.ReceiveBytes, color),
				util.PrettyBytes(peer.TransmitBytes, color))
		}

		if peer.PersistentKeepaliveInterval > 0 {
			kv["persistent keepalive"] = util.Every(peer.PersistentKeepaliveInterval, color)
		}

		fmt.Fprintln(wr)
		if _, err := t.FprintfColored(wr, color, t.Color("peer", t.Bold, t.FgYellow)+": "+t.Color("%s", t.FgYellow)+"\n", peer.PublicKey.String()); err != nil {
			return err
		}
		if _, err := t.PrintKeyValues(wr, color, "  ", kv); err != nil {
			return err
		}
	}

	return nil
}
