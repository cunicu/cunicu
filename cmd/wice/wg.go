package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/wg"
)

var (
	wgCmd = &cobra.Command{
		Use:   "wg",
		Short: "WireGuard commands",
		Args:  cobra.NoArgs,
	}

	wgGenKeyCmd = &cobra.Command{
		Use:   "genkey",
		Short: "Generates a new private key and writes it to stdout",
		RunE:  wgGenKey,
		Args:  cobra.NoArgs,
	}

	wgGenPSKCmd = &cobra.Command{
		Use:   "genpsk",
		Short: "Generates a new preshared key and writes it to stdout",
		RunE:  wgGenKey, // a preshared key is generated in the same way as a private key
		Args:  cobra.NoArgs,
	}

	wgPubKeyCmd = &cobra.Command{
		Use:   "pubkey",
		Short: "Reads a private key from stdin and writes a public key to stdout",
		RunE:  wgPubKey,
		Args:  cobra.NoArgs,
	}

	wgShowCmd = &cobra.Command{
		Use:   "show",
		Short: "Shows the current configuration and device information",
		RunE:  wgShow,
		Args:  cobra.ArbitraryArgs,
	}
)

func init() {
	rootCmd.AddCommand(wgCmd)

	wgCmd.AddCommand(wgGenKeyCmd)
	wgCmd.AddCommand(wgGenPSKCmd)
	wgCmd.AddCommand(wgPubKeyCmd)

	addClientCommand(wgCmd, wgShowCmd)
}

func wgGenKey(cmd *cobra.Command, args []string) error {
	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		os.Exit(1)
	}

	fmt.Println(key.String())

	return nil
}

func wgPubKey(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)
	privKeyStr, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read from stdin: %w", err)
	}

	privKey, err := wgtypes.ParseKey(privKeyStr)
	if err != nil {
		os.Exit(1)
	}

	fmt.Println(privKey.PublicKey().String())

	return nil
}

func wgShow(cmd *cobra.Command, args []string) error {
	sts, err := client.GetStatus(context.Background(), &pb.Empty{})
	if err != nil {
		return fmt.Errorf("failed RPC request: %w", err)
	}

	for _, intf := range sts.Interfaces {
		mdev := intf.Device()
		wdev := wg.Device(*mdev)

		if err := wdev.DumpEnv(os.Stdout); err != nil {
			return err
		}
	}

	return nil
}
