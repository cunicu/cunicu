package main

import (
	"os"
)

func main() {
	if os.Args[0] == "wg" {
		if err := wgCmd.Execute(); err != nil {
			os.Exit(1)
		}
	} else {
		if err := rootCmd.Execute(); err != nil {
			os.Exit(1)
		}
	}
}
