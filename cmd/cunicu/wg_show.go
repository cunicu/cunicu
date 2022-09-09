package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/stv0g/cunicu/pkg/wg"

	rpcproto "github.com/stv0g/cunicu/pkg/proto/rpc"
)

var (
	wgShowCmd = &cobra.Command{
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
)

func init() {
	addClientCommand(wgCmd, wgShowCmd)
}

func wgShowValidArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	comps := []string{}

	if len(args) == 0 {
		comps = []string{"all", "interfaces"}

		rpcConnect(cmd, args)
		defer rpcDisconnect(cmd, args)

		sts, err := rpcClient.GetStatus(context.Background(), &rpcproto.StatusParams{})
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

func wgShow(cmd *cobra.Command, args []string) error {
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
				cmd.Usage()
				os.Exit(1)
			}
			field = args[1]
		} else {
			field = "all"
		}
	} else {
		mode = "all"
		field = "all"
	}

	sts, err := rpcClient.GetStatus(context.Background(), &rpcproto.StatusParams{
		Intf: intf,
	})
	if err != nil {
		return fmt.Errorf("failed RPC request: %w", err)
	}

	intfNames := []string{}

	for i, intf := range sts.Interfaces {
		if mode == "interfaces" {
			intfNames = append(intfNames, intf.Name)
		} else {
			var prefix string

			mdev := intf.Device()
			wdev := wg.Device(*mdev)

			if mode == "all" {
				prefix = wdev.Name + "\t"
			}

			if field == "all" {
				if i > 0 {
					if _, err := fmt.Println(); err != nil {
						return err
					}
				}

				if err := wdev.DumpEnv(os.Stdout); err != nil {
					return err
				}
			} else {

				var value any
				switch field {
				case "public-key":
					value = wdev.PublicKey
				case "private-key":
					value = wdev.PrivateKey
				case "listen-port":
					value = wdev.ListenPort
				case "fwmark":
					value = "off"
					if wdev.FirewallMark != 0 {
						value = fmt.Sprint(wdev.FirewallMark)
					}
				}

				if value != nil {
					fmt.Printf("%s%v\n", prefix, value)
				} else {
					if field == "dump" {
						fwmark := "off"
						if wdev.FirewallMark != 0 {
							fwmark = fmt.Sprint(wdev.FirewallMark)
						}

						fmt.Printf("%s%s\t%s\t%d\t%s\n",
							prefix,
							wdev.PrivateKey,
							wdev.PublicKey,
							wdev.ListenPort,
							fwmark,
						)
					}

					for _, peer := range wdev.Peers {
						switch field {
						case "peers":
							fmt.Printf("%s%s\n", prefix, peer.PublicKey)
							continue
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
				}
			}
		}
	}

	if mode == "interfaces" {
		fmt.Println(strings.Join(intfNames, " "))
	}

	return nil
}
