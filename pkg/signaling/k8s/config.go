package k8s

import (
	"errors"
	"net/url"
)

type BackendConfig struct {
	URI *url.URL

	NodeName            string
	AnnotationOffers    string
	AnnotationPublicKey string
}

func (c *BackendConfig) Parse(uri *url.URL, options map[string]string) error {
	var ok bool

	c.URI = uri

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
