//go:build test

package cmd

import (
	"testing"
)

func TestRunMain(t *testing.T) {
	cmd.WGCmd.Execute()
}
