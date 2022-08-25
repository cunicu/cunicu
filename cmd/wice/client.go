package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"riasc.eu/wice/pkg/config"
	"riasc.eu/wice/pkg/rpc"
)

var (
	rpcClient   *rpc.Client
	rpcSockPath string
)

func addClientCommand(rcmd, cmd *cobra.Command) {
	cmd.PersistentPreRunE = connect
	cmd.PersistentPostRunE = disconnect

	pf := cmd.PersistentFlags()
	pf.StringVarP(&rpcSockPath, "rpc-socket", "s", config.DefaultSocketPath, "Unix control and monitoring socket")

	rcmd.AddCommand(cmd)
}

func connect(cmd *cobra.Command, args []string) error {
	var err error

	if rpcClient, err = rpc.Connect(rpcSockPath); err != nil {
		return fmt.Errorf("failed to connect to control socket: %w", err)
	}

	return nil
}

func disconnect(cmd *cobra.Command, args []string) error {
	return rpcClient.Close()
}