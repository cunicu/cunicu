package grpc

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type BackendConfig struct {
	URI *url.URL

	Target string

	Options []grpc.DialOption
}

func (c *BackendConfig) Parse(uri *url.URL) error {
	options := uri.Query()

	if str := options.Get("insecure"); str != "" {
		if b, err := strconv.ParseBool(str); err == nil && b {
			c.Options = append(c.Options, grpc.WithTransportCredentials(
				insecure.NewCredentials(),
			))
		}
	}

	c.URI = uri

	if uri.Host == "" {
		return errors.New("missing gRPC server url")
	}

	c.Target = fmt.Sprintf("%s", uri.Host)

	return nil
}
