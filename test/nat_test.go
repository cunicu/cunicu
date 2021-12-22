//go:build linux
// +build linux

package main_test

import (
	"testing"
)

func TestNAT(t *testing.T) {
	// var (
	// 	n   *g.Network
	// 	sw1 *g.Switch
	// 	sw2 *g.Switch
	// 	sw3 *g.Switch
	// 	b   *test.SignalingNode
	// 	r   *test.RelayNode
	// 	nl  test.NodeList

	// 	err error
	// )

	// if n, err = g.NewNetwork("", gopt.Persistent(true)); err != nil {
	// 	t.Fatalf("Failed to create network: %s", err)
	// }
	// defer n.Close()

	// if sw1, err = n.AddSwitch("sw1"); err != nil {
	// 	t.Fatalf("Failed to create switch: %s", err)
	// }

	// if sw2, err = n.AddSwitch("sw2"); err != nil {
	// 	t.Fatalf("Failed to create switch: %s", err)
	// }

	// if sw3, err = n.AddSwitch("sw"); err != nil {
	// 	t.Fatalf("Failed to create switch: %s", err)
	// }
}
