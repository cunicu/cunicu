package http

import (
	"fmt"
	"net/url"
	"time"

	"riasc.eu/wice/pkg/backend/base"
)

const (
	defaultPollInterval = 5 * time.Second
	defaultTimeout      = 10 * time.Second
)

type BackendConfig struct {
	base.BackendConfig

	PollInterval       time.Duration
	Timeout            time.Duration
	InsecureSkipVerify bool
}

func (c *BackendConfig) Parse(uri *url.URL, options map[string]string) error {
	err := c.BackendConfig.Parse(uri, options)
	if err != nil {
		return err
	}

	if skip, ok := options["insecure_skip_verify"]; ok {
		c.InsecureSkipVerify = skip == "true"
	}

	if interval, ok := options["interval"]; ok {
		c.PollInterval, err = time.ParseDuration(interval)
		if err != nil {
			return fmt.Errorf("invalid interval: %s", interval)
		}
	} else {
		c.PollInterval = defaultPollInterval
	}

	if timeout, ok := options["timeout"]; ok {
		c.Timeout, err = time.ParseDuration(timeout)
		if err != nil {
			return fmt.Errorf("invalid timeout: %s", timeout)
		}
	} else {
		c.Timeout = defaultTimeout
	}

	return nil
}
