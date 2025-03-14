// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"net"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"cunicu.li/cunicu/pkg/crypto"
	rpcproto "cunicu.li/cunicu/pkg/proto/rpc"
	"cunicu.li/cunicu/pkg/wg"
)

type inviteOptions struct {
	listenPort int
}

func init() { //nolint:gochecknoinits
	opts := &inviteOptions{}
	cmd := &cobra.Command{
		Use:   "invite [interface]",
		Short: "Add a new peer to the local daemon configuration and return the required configuration for this new peer",
		Run: func(cmd *cobra.Command, args []string) {
			invite(cmd, args, opts)
		},
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: interfaceValidArgs,
	}

	addClientCommand(rootCmd, cmd)

	pf := cmd.PersistentFlags()

	pf.IntVarP(&opts.listenPort, "listen-port", "L", wg.DefaultPort, "Listen port for generated config")
}

func invite(_ *cobra.Command, args []string, opts *inviteOptions) {
	sk, err := crypto.GeneratePrivateKey()
	if err != nil {
		logger.Fatal("Failed to generate private key", zap.Error(err))
	}

	intf := args[0]

	// First add a new peer to the running daemon via runtime configuration RPCs
	addPeerResp, err := rpcClient.AddPeer(context.Background(), &rpcproto.AddPeerParams{
		Interface: intf,
		PublicKey: sk.PublicKey().Bytes(),
	})
	if err != nil {
		logger.Fatal("Failed to add new peer to daemon", zap.Error(err))
	}

	if true {
		// Generate a wg-quick configuration
		mtu := int(addPeerResp.Interface.Mtu)
		pk, _ := crypto.ParseKeyBytes(addPeerResp.Interface.PublicKey)

		cfgPeer := wgtypes.PeerConfig{
			PublicKey: wgtypes.Key(pk),
		}

		cfg := wg.Config{
			Config: wgtypes.Config{
				PrivateKey: (*wgtypes.Key)(&sk),
				ListenPort: &opts.listenPort,
				Peers:      []wgtypes.PeerConfig{cfgPeer},
			},
			Address: []net.IPNet{},
			MTU:     &mtu,
		}

		if addPeerResp.Invitation.Endpoint != "" {
			cfg.PeerEndpoints = []string{addPeerResp.Invitation.Endpoint}
		}

		if err := cfg.Dump(os.Stdout); err != nil {
			logger.Fatal("Failed to dump config", zap.Error(err))
		}
	}
}
