package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"riasc.eu/wice/pkg/crypto"
)

var (
	v4, v6 bool

	addressesCmd = &cobra.Command{
		Use:    "addresses",
		Short:  "Calculate link-local IPv4 and IPv6 addresses from a WireGuard X25519 public key",
		Run:    addresses,
		Hidden: true,
	}
)

func init() {
	pf := addressesCmd.PersistentFlags()
	pf.BoolVarP(&v4, "ipv4", "4", false, "Print IPv4 address")
	pf.BoolVarP(&v6, "ipv6", "6", false, "Print IPv6 address")

	rootCmd.AddCommand(addressesCmd)
}

func addresses(cmd *cobra.Command, args []string) {
	logger := zap.L()

	keyB64, err := io.ReadAll(os.Stdin)
	if err != nil {
		logger.Fatal("Failed to read from stdin", zap.Error(err))
	}

	key, err := crypto.ParseKey(string(keyB64))
	if err != nil {
		logger.Fatal("Failed to parse key",
			zap.Error(err),
			zap.String("key", string(keyB64)))
	}

	if v6 || (!v4 && !v6) {
		fmt.Printf("%s\n", key.IPv6Address())
	}

	if v4 || (!v4 && !v6) {
		fmt.Printf("%s\n", key.IPv4Address())
	}
}
