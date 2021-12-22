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

var defaultConfig = BackendConfig{
	AnnotationOffers:    defaultAnnotationOffers,
	AnnotationPublicKey: defaultAnnotationPublicKey,
}

func (c *BackendConfig) Parse(uri *url.URL) error {
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

	uri.RawQuery = ""
	c.URI = uri

	return nil
}
