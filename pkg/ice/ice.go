// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package ice extends the pion/ice package with custom (un-)marshaling support
package ice

import (
	"net/url"
	"strings"

	"github.com/pion/stun/v3"
)

func ParseURL(urlStr string) (*stun.URI, string, string, url.Values, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, "", "", nil, err
	}

	user := ""
	pass := ""

	if strings.Contains(u.Opaque, "@") {
		op := strings.Split(u.Opaque, "@")
		up := strings.Split(op[0], ":")
		user = up[0]

		if len(up) > 1 {
			pass = up[1]
		}

		u.Opaque = op[1]
	}

	q := u.Query()
	if t := q.Get("transport"); t != "" {
		u.RawQuery = "transport=" + t
	} else {
		u.RawQuery = ""
	}

	iu, err := stun.ParseURI(u.String())
	if err != nil {
		return nil, "", "", nil, err
	}

	return iu, user, pass, q, nil
}
