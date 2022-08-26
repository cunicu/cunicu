//go:build tracer

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"riasc.eu/wice/pkg/util"
	"riasc.eu/wice/pkg/wg/tracer"
)

var (
	wgExtractHandshakesCmd = &cobra.Command{
		Use:   "extract-handshakes",
		Short: "Extract WireGuard handhshakes from Linux kernel",
		Long:  "This command extracts ephemeral session secrets from handshakes of local WireGuard interfaces via Linux eBPF and kProbes and logs them to the standard output in the keylog format used by Wireshark",
		RunE:  wgExtractHandshakes,
	}
)

func init() {
	wgCmd.AddCommand(wgExtractHandshakesCmd)
}

func wgExtractHandshakes(cmd *cobra.Command, args []string) error {
	logger := zap.L().Named("tracer")

	ht, err := tracer.NewHandshakeTracer()
	if err != nil {
		logger.Fatal("Failed to create tracer", zap.Error(err))
	}

	sigs := util.SetupSignals()

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
