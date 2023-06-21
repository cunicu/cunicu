// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package log

import "github.com/onsi/ginkgo/v2"

type ginkgoSyncWriter struct {
	ginkgo.GinkgoWriterInterface
}

func (w *ginkgoSyncWriter) Close() error {
	return nil
}

func (w *ginkgoSyncWriter) Sync() error {
	return nil
}
