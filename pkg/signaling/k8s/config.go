package k8s

import (
	"errors"
	"net/url"
)

type BackendConfig struct {
	signaling.BackendConfig

	Kubeconfig   string
	NodeName            string
	AnnotationOffers    string
	AnnotationPublicKey string
}

var defaultConfig = BackendConfig{
	AnnotationOffers:    defaultAnnotationOffers,
	AnnotationPublicKey: defaultAnnotationPublicKey,
}

func (c *BackendConfig) Parse(cfg *signaling.BackendConfig) error {
	c.BackendConfig = *cfg

	options := uri.Query()

	if str := options.Get("node"); str == "" {
		return errors.New("missing backend option: node")
	}

	if str := options.Get("annotation-offers"); str != "" {
		c.AnnotationOffers = str
	}

	if str := options.Get("annotation-public-key"); str != "" {
		c.AnnotationPublicKey = str
	}

	c.Kubeconfig = c.URI.Path

	return nil
}
