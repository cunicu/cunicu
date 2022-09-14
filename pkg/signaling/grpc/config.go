package grpc

import (
	"fmt"

	"google.golang.org/grpc"

	"github.com/stv0g/cunicu/pkg/signaling"
)

type BackendConfig struct {
	signaling.BackendConfig

	Target  string
	Options []grpc.DialOption
}

func (c *BackendConfig) Parse(cfg *signaling.BackendConfig) error {
	var err error

	c.BackendConfig = *cfg
	c.Target, c.Options, err = ParseURL(c.URI.String())
	if err != nil {
		return fmt.Errorf("failed to parse gRPC URL:%w", err)
	}

	return nil
}
