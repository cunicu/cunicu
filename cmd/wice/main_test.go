// go:build testmain
//go:build testmain
// +build testmain

package main_test

import (
	"os"
	"testing"

	m "riasc.eu/wice/cmd/wice"

	"golang.org/x/exp/slices"
)

func TestMain(t *testing.T) {
	i := slices.Index(os.Args, "--")
	if i >= 0 {
		os.Args = append([]string{os.Args[0]}, os.Args[i+1:]...)
	}

	m.RootCmd.Execute()
}
