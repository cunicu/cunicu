// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/log"
)

type addressesOptions struct {
	mask bool
}

//nolint:gochecknoinits
func init() {
	opts := &addressesOptions{}
	cmd := &cobra.Command{
		Use:   "addresses",
		Short: "Derive IPv4 and IPv6 addresses from a WireGuard X25519 public key",
		Long: `cunīcu auto-configuration feature derives and assigns IPv4 and IPv6 addresses based on the public key of the WireGuard interface.
This sub-command accepts a WireGuard public key on the standard input and prints out the calculated IP addresses on the standard output.
`,
		Run: func(cmd *cobra.Command, args []string) {
			addresses(cmd, args, opts)
		},
		Example: `$ wg genkey | wg pubkey | cunicu addresses
fc2f:9a4d:777f:7a97:8197:4a5d:1d1b:ed79
10.237.119.127`,
		Args:              cobra.ArbitraryArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
	}

	pf := cmd.PersistentFlags()
	pf.BoolVarP(&opts.mask, "mask", "m", false, "Print CIDR mask")

	rootCmd.AddCommand(cmd)
}

func addresses(_ *cobra.Command, args []string, opts *addressesOptions) {
	logger := log.Global

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

	if len(args) == 0 {
		args = config.DefaultPrefixes
	}

	for _, ps := range args {
		_, p, err := net.ParseCIDR(ps)
		if err != nil {
			logger.Fatal("Failed to parse prefix", zap.Error(err))
		}

		q := key.IPAddress(*p)

		if opts.mask {
			fmt.Println(q.String())
		} else {
			fmt.Println(q.IP.String())
		}
	}
}
