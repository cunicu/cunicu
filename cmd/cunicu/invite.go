package main

import (
	"bytes"
	"context"
	"net"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/util/terminal"
	"github.com/stv0g/cunicu/pkg/wg"

	rpcproto "github.com/stv0g/cunicu/pkg/proto/rpc"
)

var inviteCmd = &cobra.Command{
	Use:               "invite [interface]",
	Short:             "Add a new peer to the local daemon configuration and return the required configuration for this new peer",
	Run:               invite,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: interfaceValidArgs,
}

var listenPort int
var qrCode bool

func init() {
	addClientCommand(rootCmd, inviteCmd)

	pf := inviteCmd.PersistentFlags()

	pf.IntVarP(&listenPort, "listen-port", "L", wg.DefaultPort, "Listen port for generated config")
	pf.BoolVarP(&qrCode, "qr-code", "Q", false, "Show config as QR code in terminal")
}

func invite(cmd *cobra.Command, args []string) {
	sk, err := crypto.GeneratePrivateKey()
	if err != nil {
		logger.Fatal("Failed to generate private key", zap.Error(err))
	}

	intf := args[0]

	// First add a new new peer to the running daemon via runtime configuration RPCs
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
				ListenPort: &listenPort,
				Peers:      []wgtypes.PeerConfig{cfgPeer},
			},
			Address: []net.IPNet{},
			MTU:     &mtu,
		}

		if addPeerResp.Invitation.Endpoint != "" {
			cfg.PeerEndpoints = []string{addPeerResp.Invitation.Endpoint}
		}

		if qrCode {
			buf := &bytes.Buffer{}
			if err := cfg.Dump(buf); err != nil {
				logger.Fatal("Failed to dump config", zap.Error(err))
			}

			terminal.QRCode(buf.String())
		} else {
			if err := cfg.Dump(os.Stdout); err != nil {
				logger.Fatal("Failed to dump config", zap.Error(err))
			}
		}
	}
}
