// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func init() { //nolint:gochecknoinits
	genKeyCmd := &cobra.Command{
		Use:   "genkey",
		Short: "Generates a random private key in base64 and prints it to standard output.",
		RunE:  wgGenKey,
		Args:  cobra.NoArgs,
	}

	genPSKCmd := &cobra.Command{
		Use:   "genpsk",
		Short: "Generates a random preshared key in base64 and prints it to standard output.",
		RunE:  wgGenKey, // a preshared key is generated in the same way as a private key
		Args:  cobra.NoArgs,
	}

	pubKeyCmd := &cobra.Command{
		Use:   "pubkey",
		Short: "Calculates a public key and prints it in base64 to standard output.",
		Long:  `Calculates a public key and prints it in base64 to standard output from a corresponding private key (generated with genkey) given in base64 on standard input.`,
		Example: `# A private key and a corresponding public key may be generated at once by calling:
$ umask 077
$ wg genkey | tee private.key | wg pubkey > public.key`,
		RunE: wgPubKey,
		Args: cobra.NoArgs,
	}

	wgCmd.AddCommand(genKeyCmd)
	wgCmd.AddCommand(genPSKCmd)
	wgCmd.AddCommand(pubKeyCmd)
}

func wgGenKey(_ *cobra.Command, _ []string) error {
	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		os.Exit(1)
	}

	fmt.Printf("%s\n", key)

	return nil
}

func wgPubKey(_ *cobra.Command, _ []string) error {
	privKeyStrBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read from stdin: %w", err)
	}

	privKey, err := wgtypes.ParseKey(string(privKeyStrBytes))
	if err != nil {
		os.Exit(1)
	}

	fmt.Printf("%s\n", privKey.PublicKey())

	return nil
}
