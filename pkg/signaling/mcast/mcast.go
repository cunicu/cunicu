// SPDX-FileCopyrightText: 2025 Adam Rizkalla <ajarizzo@gmail.com>
// SPDX-License-Identifier: Apache-2.0

// Package mcast implements a signaling backend using multicast
package mcast

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
)

var errInvalidAddress = errors.New("missing multicast address")

func ParseURL(urlStr string) (string, BackendOptions, error) {

	o := BackendOptions{
		Interface: nil,
		Loopback:  false,
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return "", o, err
	}

	q := u.Query()

	if q.Has("interface") {
		if o.Interface, err = net.InterfaceByName(q.Get("interface")); err != nil {
			return "", o, fmt.Errorf("failed to parse 'interface' option: %w", err)
		}
	}

	if q.Has("loopback") {
		var err error
		if o.Loopback, err = strconv.ParseBool(q.Get("loopback")); err != nil {
			return "", o, fmt.Errorf("failed to parse 'loopback' option: %w", err)
		}
	}

	if u.Host == "" {
		return "", o, errInvalidAddress
	}

	return u.Host, o, nil
}
