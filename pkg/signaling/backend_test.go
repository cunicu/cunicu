// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package signaling_test

import (
	"net/url"

	"github.com/stv0g/cunicu/pkg/signaling"
	"github.com/stv0g/cunicu/pkg/signaling/inprocess"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type readyHandler struct {
	ready chan any
}

func (h *readyHandler) OnSignalingBackendReady(_ signaling.Backend) {
	close(h.ready)
}

var _ = It("can create new a new backend", func() {
	uri, err := url.Parse("inprocess:")
	Expect(err).To(Succeed(), "Failed to parse URL: %s", err)

	h := &readyHandler{make(chan any)}
	cfg := &signaling.BackendConfig{
		URI:     uri,
		OnReady: []signaling.BackendReadyHandler{h},
	}

	b, err := signaling.NewBackend(cfg)
	Expect(err).To(Succeed(), "Failed to create new backend: %s", err)

	_, isInprocessBackend := b.(*inprocess.Backend)
	Expect(isInprocessBackend).To(BeTrue())

	// Wait until the backend is ready
	Eventually(h.ready).Should(BeClosed())
})
