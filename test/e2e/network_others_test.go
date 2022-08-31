//go:build !tracer

package e2e_test

type HandshakeTracer any

func (n *Network) StartHandshakeTracer() {}
func (n *Network) StopHandshakeTracer()  {}
