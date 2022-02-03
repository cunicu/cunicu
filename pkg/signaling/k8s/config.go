package k8s

import (
	"riasc.eu/wice/pkg/signaling"
)

type BackendConfig struct {
	signaling.BackendConfig

	Kubeconfig   string
	Namespace    string
	GenerateName string
}

var defaultConfig = BackendConfig{
	GenerateName: "wice-",
	Namespace:    "riasc-system",
}

func (c *BackendConfig) Parse(cfg *signaling.BackendConfig) error {
	c.BackendConfig = *cfg

	c.Kubeconfig = c.URI.Path

	return nil
}
