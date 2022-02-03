package grpc

import (
	"errors"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"riasc.eu/wice/pkg/signaling"
)

type BackendConfig struct {
	signaling.BackendConfig

	Target string

	Options []grpc.DialOption
}

func (c *BackendConfig) Parse(cfg *signaling.BackendConfig) error {
	c.BackendConfig = *cfg

	options := c.URI.Query()
	if str := options.Get("insecure"); str != "" {
		if b, err := strconv.ParseBool(str); err == nil && b {
			c.Options = append(c.Options, grpc.WithTransportCredentials(
				insecure.NewCredentials(),
			))
		}
	}

	if c.URI.Host == "" {
		return errors.New("missing gRPC server url")
	}

	c.Target = c.URI.Host

	return nil
}
