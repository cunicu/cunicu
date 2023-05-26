// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build !(linux || freebsd || darwin)

package link

func CreateWireGuardLink(_ string) (Link, error) {
	return nil, errNotSupported
}

func FindLink(_ string) (Link, error) {
	return nil, errNotSupported
}
