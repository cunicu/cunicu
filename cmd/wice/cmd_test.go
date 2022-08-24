//go:build test

package main

import (
	"testing"
)

func TestRunMain(t *testing.T) {
	wgCmd.Execute()
}
