package signaling_test

import (
	"net/url"
	"testing"

	"riasc.eu/wice/internal/log"
	"riasc.eu/wice/pkg/signaling"
	"riasc.eu/wice/pkg/signaling/inprocess"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Signaling Suite")
}

var _ = It("can create new a new backend", func() {
	uri, err := url.Parse("inprocess:")
	Expect(err).To(Succeed(), "Failed to parse URL: %s", err)

	events := log.NewEventLogger()

	cfg := &signaling.BackendConfig{
		URI: uri,
	}

	b, err := signaling.NewBackend(cfg, events)
	Expect(err).To(Succeed(), "Failed to create new backend: %s", err)

	_, isInprocessBackend := b.(*inprocess.Backend)
	Expect(isInprocessBackend).To(BeTrue())
})
