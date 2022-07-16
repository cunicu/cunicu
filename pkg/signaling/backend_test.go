package signaling_test

import (
	"net/url"
	"sync/atomic"

	"riasc.eu/wice/pkg/signaling"
	"riasc.eu/wice/pkg/signaling/inprocess"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type readyHandler int32

func (h *readyHandler) OnBackendReady(b signaling.Backend) {
	atomic.AddInt32((*int32)(h), 1)
}

var _ = It("can create new a new backend", func() {
	uri, err := url.Parse("inprocess:")
	Expect(err).To(Succeed(), "Failed to parse URL: %s", err)

	h := readyHandler(0)

	cfg := &signaling.BackendConfig{
		URI: uri,
	}

	b, err := signaling.NewBackend(cfg)
	Expect(err).To(Succeed(), "Failed to create new backend: %s", err)

	b.OnReady(&h)

	_, isInprocessBackend := b.(*inprocess.Backend)
	Expect(isInprocessBackend).To(BeTrue())

	// Wait until the backend is ready
	Eventually(func() int32 {
		return atomic.LoadInt32((*int32)(&h))
	}).ShouldNot(BeZero())
})
