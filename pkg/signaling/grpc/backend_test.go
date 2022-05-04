package grpc_test

import (
	"net"
	"testing"

	"riasc.eu/wice/internal/test"
	"riasc.eu/wice/pkg/signaling/grpc"
)

func TestBackend(t *testing.T) {
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		t.Fatalf("Failed to listen: %s", err)
	}

	// Start local dummy gRPC server
	svr := grpc.NewServer()
	go svr.Serve(l)
	defer svr.Stop()

	test.TestBackend(t, "grpc://127.0.0.1:8080?insecure=true", 10)
}
