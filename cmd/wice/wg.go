package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var (
	wgCmd = &cobra.Command{
		Use:   "wg",
		Short: "Wireguard commands",
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
)

func init() {
	rootCmd.AddCommand(wgCmd)

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
	reader := bufio.NewReader(os.Stdin)
	privKeyStr, _ := reader.ReadString('\n')

	privKey, err := wgtypes.ParseKey(privKeyStr)
	if err != nil {
		os.Exit(1)
	}

	fmt.Println(privKey.PublicKey().String())

	return nil
}
