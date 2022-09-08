package k8s

import (
	"github.com/stv0g/cunicu/pkg/signaling"
)

type BackendConfig struct {
	signaling.BackendConfig

	Kubeconfig   string
	Namespace    string
	GenerateName string
}

var defaultConfig = BackendConfig{
	GenerateName: "wice-",
	Namespace:    "wice",
}

func (c *BackendConfig) Parse(cfg *signaling.BackendConfig) error {
	c.BackendConfig = *cfg

	c.Kubeconfig = c.URI.Path

	if ns := c.URI.Query().Get("namespace"); ns != "" {
		c.Namespace = ns
	}

	return nil
}
