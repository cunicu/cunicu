package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"riasc.eu/wice/internal"
	"riasc.eu/wice/internal/cli"
)

var (
	// Used for flags.
	cfgFile  string
	sockPath string

	rootCmd = &cobra.Command{
		Use:   "wicectl",
		Short: "A client tool to control the WICE daemon",
	}
)

func init() {
	cobra.OnInitialize(
		internal.SetupRand,
		setupLogging,
		setupConfig,
	)

	pf := rootCmd.PersistentFlags()

	pf.StringVar(&sockPath, "socket", "/var/run/wice.sock", "Unix control and monitoring socket")
	pf.StringVar(&cfgFile, "config", "", "Path to config file (default $HOME/.wice.yaml)")

	viper.BindPFlag("socket", pf.Lookup("socket"))

	rootCmd.AddCommand(shutdownCmd)
	rootCmd.AddCommand(monitorCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(cli.NewDocsCommand(rootCmd))
}

func setupLogging() {
	logger = internal.SetupLogging()
}

func setupConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".wice.yaml")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		logger.Debug("Using config file", zap.String("file", viper.ConfigFileUsed()))
	}
}
