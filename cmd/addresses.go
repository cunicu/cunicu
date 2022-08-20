package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"riasc.eu/wice/pkg/crypto"
)

var (
	addressesCmd = &cobra.Command{
		Use:    "addresses",
		Short:  "Calculate the IPv6 link-local address from a public key",
		Run:    addresses,
		Hidden: true,
	}
)

func init() {
	RootCmd.AddCommand(addressesCmd)
}

func addresses(cmd *cobra.Command, args []string) {
	logger := zap.L()

	reader := bufio.NewReader(os.Stdin)
	keyB64, err := reader.ReadString('\n')
	if err != nil {
		logger.Fatal("Failed to read from stdin", zap.Error(err))
	}

	key, err := crypto.ParseKey(keyB64)
	if err != nil {
		logger.Fatal("Failed to parse key", zap.Error(err), zap.String("key", keyB64))
	}

	fmt.Println(key.IPv6Address().String())
	fmt.Println(key.IPv4Address().String())
}
