// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//nolint:goconst
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	proto "github.com/stv0g/cunicu/pkg/proto/core"
	rpcproto "github.com/stv0g/cunicu/pkg/proto/rpc"
	"github.com/stv0g/cunicu/pkg/wg"
)

var errUnknownField = errors.New("unknown field")

func init() { //nolint:gochecknoinits
	cmd := &cobra.Command{
		Use:   "show { interface-name | all | interfaces } [{ public-key | private-key | listen-port | fwmark | peers | preshared-keys | endpoints | allowed-ips | latest-handshakes | transfer | persistent-keepalive | dump }]",
		Short: "Shows current WireGuard configuration and runtime information of specified [interface].",
		Long: `Shows current WireGuard configuration and runtime information of specified [interface].
		
If no [interface] is specified, [interface] defaults to 'all'.

If 'interfaces' is specified, prints a list of all WireGuard interfaces, one per line, and quits.

If no options are given after the interface specification, then prints a list of all attributes in a visually pleasing way meant for the terminal.
Otherwise, prints specified information grouped by newlines and tabs, meant to be used in scripts.

For this script-friendly display, if 'all' is specified, then the first field for all categories of information is the interface name.

If 'dump' is specified, then several lines are printed; the first contains in order separated by tab: private-key, public-key, listen-port, fwmark.
Subsequent lines are printed for each peer and contain in order separated by tab: public-key, preshared-key, endpoint, allowed-ips, latest-handshake, transfer-rx, transfer-tx, persistent-keepalive.`,
		RunE:              wgShow,
		Args:              cobra.MaximumNArgs(2),
		ValidArgsFunction: wgShowValidArgs,
	}

	addClientCommand(wgCmd, cmd)
}

func wgShowValidArgs(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	comps := []string{}

	if len(args) == 0 {
		comps = []string{"all", "interfaces"}

		if err := rpcConnect(cmd, args); err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		defer rpcDisconnect(cmd, args) //nolint:errcheck

		sts, err := rpcClient.GetStatus(context.Background(), &rpcproto.GetStatusParams{})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		for _, i := range sts.Interfaces {
			comps = append(comps, i.Name)
		}
	} else if len(args) == 1 {
		comps = []string{"public-key", "private-key", "listen-port", "fwmark", "peers", "preshared-keys", "endpoints", "allowed-ips", "latest-handshakes", "transfer", "persistent-keepalive", "dump"}
	}

	return comps, cobra.ShellCompDirectiveNoFileComp
}

func wgShow(_ *cobra.Command, args []string) error {
	intf, mode, field, err := parseWgShowArgs(args)
	if err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	sts, err := rpcClient.GetStatus(context.Background(), &rpcproto.GetStatusParams{
		Interface: intf,
	})
	if err != nil {
		return fmt.Errorf("failed RPC request: %w", err)
	}

	switch mode {
	case "interfaces":
		showInterfaceList(sts.Interfaces)
	default:
		for i, intf := range sts.Interfaces {
			if err := showInterfaceDetails(intf.Device(), i, mode, field); err != nil {
				return err
			}
		}
	}

	return nil
}

func parseWgShowArgs(args []string) (string, string, string, error) {
	var intf, mode, field string

	if len(args) > 0 {
		switch args[0] {
		case "all":
			mode = args[0]
		case "interfaces":
			mode = args[0]
		default:
			mode = "interface"
			intf = args[0]
		}

		if len(args) > 1 {
			switch args[1] {
			case "public-key":
			case "private-key":
			case "listen-port":
			case "fwmark":
			case "peers":
			case "preshared-keys":
			case "endpoints":
			case "allowed-ips":
			case "latest-handshakes":
			case "transfer":
			case "persistent-keepalive":
			case "dump":
			default:
				return "", "", "", fmt.Errorf("%w: %s", errUnknownField, args[1])
			}
			field = args[1]
		} else {
			field = "all"
		}
	} else {
		mode = "all"
		field = "all"
	}

	return intf, mode, field, nil
}

func showInterfaceList(intfs []*proto.Interface) {
	intfNames := []string{}

	for _, intf := range intfs {
		intfNames = append(intfNames, intf.Name)
	}

	fmt.Println(strings.Join(intfNames, " "))
}

func showInterfaceDetails(dev *wg.Interface, i int, mode, field string) error {
	var prefix string

	if mode == "all" {
		prefix = dev.Name + "\t"
	}

	if field == "all" {
		if i > 0 {
			if _, err := fmt.Println(); err != nil {
				return err
			}
		}

		if err := dev.DumpEnv(os.Stdout); err != nil {
			return err
		}
	} else {
		var value any
		switch field {
		case "public-key":
			value = dev.PublicKey
		case "private-key":
			value = dev.PrivateKey
		case "listen-port":
			value = dev.ListenPort
		case "fwmark":
			value = "off"
			if dev.FirewallMark != 0 {
				value = fmt.Sprint(dev.FirewallMark)
			}
		}

		if value != nil {
			fmt.Printf("%s%v\n", prefix, value)
		} else {
			if field == "dump" {
				fwmark := "off"
				if dev.FirewallMark != 0 {
					fwmark = fmt.Sprint(dev.FirewallMark)
				}

				fmt.Printf("%s%s\t%s\t%d\t%s\n",
					prefix,
					dev.PrivateKey,
					dev.PublicKey,
					dev.ListenPort,
					fwmark,
				)
			}

			for _, peer := range dev.Peers {
				peer := peer

				if field == "peers" {
					fmt.Printf("%s%s\n", prefix, peer.PublicKey)
				} else {
					showPeerDetails(&peer, field, prefix)
				}
			}
		}
	}

	return nil
}

func showPeerDetails(peer *wgtypes.Peer, field, prefix string) {
	var value any

	switch field {
	case "preshared-keys":
		value = peer.PresharedKey
	case "endpoints":
		value = peer.Endpoint
	case "allowed-ips":
		aips := []string{}
		for _, aip := range peer.AllowedIPs {
			aips = append(aips, aip.String())
		}
		value = strings.Join(aips, " ")
	case "latest-handshakes":
		value = peer.LastHandshakeTime.Unix()
		if peer.LastHandshakeTime.IsZero() {
			value = 0
		}
	case "transfer":
		value = fmt.Sprintf("%d\t%d", peer.ReceiveBytes, peer.TransmitBytes)
	case "persistent-keepalive":
		value = peer.PersistentKeepaliveInterval.Seconds()
		if peer.PersistentKeepaliveInterval == 0 {
			value = "off"
		}
	case "dump":
		as := []string{}
		for _, aip := range peer.AllowedIPs {
			as = append(as, aip.String())
		}
		aIPs := strings.Join(as, ",")

		zero := wgtypes.Key{}
		psk := "(none)"
		if peer.PresharedKey != zero {
			psk = peer.PresharedKey.String()
		}

		ep := ""
		if peer.Endpoint != nil {
			ep = peer.Endpoint.String()
		}

		pka := "off"
		if peer.PersistentKeepaliveInterval.Seconds() > 0 {
			pka = fmt.Sprintf("%d", int(peer.PersistentKeepaliveInterval.Seconds()))
		}

		lhs := int64(0)
		if !peer.LastHandshakeTime.IsZero() {
			lhs = peer.LastHandshakeTime.Unix()
		}

		value = fmt.Sprintf("%s\t%s\t%s\t%s\t%d\t%d\t%d\t%s",
			peer.PublicKey,
			psk, ep, aIPs, lhs,
			peer.ReceiveBytes,
			peer.TransmitBytes,
			pka,
		)
	}

	fmt.Printf("%s%s\t%v\n", prefix, peer.PublicKey, value)
}
