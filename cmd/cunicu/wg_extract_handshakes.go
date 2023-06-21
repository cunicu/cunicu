// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build tracer

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	osx "github.com/stv0g/cunicu/pkg/os"
	"github.com/stv0g/cunicu/pkg/wg/tracer"
	"go.uber.org/zap"
)

func init() {
	cmd := &cobra.Command{
		Use:   "extract-handshakes",
		Short: "Extract WireGuard handshakes from Linux kernel",
		Long: `This command extracts ephemeral session secrets from handshakes of local WireGuard interfaces via Linux eBPF and kProbes.
The extracted keys are logged to the standard output in the keylog format used by Wireshark.

See: https://wiki.wireshark.org/WireGuard#key-log-format
`,
		RunE: wgExtractHandshakes,
	}

	wgCmd.AddCommand(cmd)
}

func wgExtractHandshakes(cmd *cobra.Command, args []string) error {
	logger := log.Global.Named("tracer")

	ht, err := tracer.NewHandshakeTracer()
	if err != nil {
		logger.Fatal("Failed to create tracer", zap.Error(err))
	}

	sigs := osx.SetupSignals()

out:
	for {
		select {
		case <-sigs:
			break out

		case err := <-ht.Errors:
			logger.Fatal("Tracer error", zap.Error(err))

		case hs := <-ht.Handshakes:
			fmt.Fprintf(os.Stderr, "=== New Handshake at %s\n", hs.Time())
			hs.DumpKeyLog(stdout)
		}
	}

	if err := ht.Close(); err != nil {
		log.Fatalf("Failed to close tracer: %s", err)
	}

	return nil
}
