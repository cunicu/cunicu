// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package hsync

import (
	"errors"
	"net"
	"strings"
)

var (
	errMissingNames     = errors.New("missing names")
	errInvalidIPAddress = errors.New("failed to parse IP address")
)

type Host struct {
	IP      net.IP
	Names   []string
	Comment string
}

func ParseHost(line string) (Host, error) {
	tokenStrs := strings.Split(line, "#")
	ipNameStrs := strings.Fields(tokenStrs[0])

	h := Host{}

	if len(tokenStrs) > 1 {
		h.Comment = strings.TrimSpace(tokenStrs[1])
	}

	if len(ipNameStrs) > 1 {
		h.Names = ipNameStrs[1:]
	} else {
		return h, errMissingNames
	}

	if h.IP = net.ParseIP(ipNameStrs[0]); h.IP == nil {
		return h, errInvalidIPAddress
	}

	return h, nil
}

func (h *Host) Line() (string, error) {
	parts := []string{
		h.IP.String(),
	}

	parts = append(parts, h.Names...)
	if h.Comment != "" {
		parts = append(parts, "#", h.Comment)
	}

	return strings.Join(parts, " "), nil
}
