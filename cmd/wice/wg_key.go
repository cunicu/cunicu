package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var (
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
)

func init() {
	wgCmd.AddCommand(wgGenKeyCmd)
	wgCmd.AddCommand(wgGenPSKCmd)
	wgCmd.AddCommand(wgPubKeyCmd)
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
	privKeyStrBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read from stdin: %w", err)
	}

	privKey, err := wgtypes.ParseKey(string(privKeyStrBytes))
	if err != nil {
		os.Exit(1)
	}

	fmt.Println(privKey.PublicKey().String())

	return nil
}
