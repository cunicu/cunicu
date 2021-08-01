package base

import "net/url"

type BackendConfig struct {
	URI *url.URL
}

func (c *BackendConfig) Parse(uri *url.URL, options map[string]string) error {
	c.URI = uri

	return nil
}
