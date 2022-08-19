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
	traceHandshakeCmd = &cobra.Command{
		Use:    "trace_handshakes",
		Short:  "Trace WireGuard handhshakes via Linux Kprobes",
		Long:   "This command traces handshakes of local WireGuard interfaces via Linux Kprobes in order to extract static, pre-shared and ephemeral keys and logs them to the standard output in the keylog format used by Wireshark",
		RunE:   traceHandshakes,
		Hidden: true,
	}
)

func init() {
	RootCmd.AddCommand(traceHandshakeCmd)
}

func traceHandshakes(cmd *cobra.Command, args []string) error {
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
			hs.DumpKeyLog(os.Stdout)
		}
	}

	if err := ht.Close(); err != nil {
		log.Fatalf("Failed to close tracer: %s", err)
	}

	return nil
}
