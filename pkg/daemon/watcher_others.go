// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build !linux

package daemon

func (w *Watcher) watchKernelInterfaces() error {
	return errNotSupported
}
