package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/stv0g/cunicu/pkg/crypto"
	"go.uber.org/zap"
)

var (
	v4, v6 bool

	addressesCmd = &cobra.Command{
		Use:   "addresses",
		Short: "Calculate link-local IPv4 and IPv6 addresses from a WireGuard X25519 public key",
		Long: `cunīcu auto-configuration feature derives and assigns link-local IPv4 and IPv6 addresses based on the public key of the WireGuard interface.
This sub-command accepts a WireGuard public key on the standard input and prints out the calculated IP addresses on the standard output.
`,
		Run: addresses,
		Example: `$ wg genkey | wg pubkey | cunicu addresses
fe80::e3be:9673:5a98:9348/64
169.254.29.188/16`,
	}
)

func init() {
	pf := addressesCmd.PersistentFlags()
	pf.BoolVarP(&v4, "ipv4", "4", false, "Print IPv4 address only")
	pf.BoolVarP(&v6, "ipv6", "6", false, "Print IPv6 address only")

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
