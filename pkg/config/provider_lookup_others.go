// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build !windows

package config

import "github.com/miekg/dns"

const resolvConfPath = "/etc/resolv.conf"

func dnsClientConfig() (*dns.ClientConfig, error) {
	return dns.ClientConfigFromFile(resolvConfPath)
}
