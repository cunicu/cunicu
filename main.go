package main

import (
	"os"

	"riasc.eu/wice/cmd"
)

func main() {
	if os.Args[0] == "wg" {
		if err := cmd.WGCmd.Execute(); err != nil {
			os.Exit(1)
		}
	} else {
		if err := cmd.RootCmd.Execute(); err != nil {
			os.Exit(1)
		}
	}
}
