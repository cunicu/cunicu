//go:build !tracer

package test_test

type HandshakeTracer any

func (n *Network) StartHandshakeTracer() {}
func (n *Network) StopHandshakeTracer()  {}
