package k8s

import (
	"errors"
	"net/url"

	"riasc.eu/wice/pkg/backend/base"
)

type BackendConfig struct {
	base.BackendConfig

	NodeName            string
	AnnotationOffers    string
	AnnotationPublicKey string
}

func (c *BackendConfig) Parse(uri *url.URL, options map[string]string) error {
	var ok bool

	err := c.BackendConfig.Parse(uri, options)
	if err != nil {
		return err
	}

	c.NodeName, ok = options["nodename"]
	if !ok {
		return errors.New("missing backend option: nodename")
	}

	c.AnnotationOffers, ok = options["annotation-offers"]
	if !ok {
		c.AnnotationOffers = defaultAnnotationOffers
	}

	c.AnnotationPublicKey, ok = options["annotation-public-key"]
	if !ok {
		c.AnnotationPublicKey = defaultAnnotationPublicKey
	}

	return nil
}
