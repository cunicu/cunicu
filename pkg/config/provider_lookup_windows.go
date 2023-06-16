// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"strings"
	"unsafe"

	"github.com/miekg/dns"
	"golang.org/x/sys/windows"
)

func dnsClientConfig() (*dns.ClientConfig, error) {
	l := uint32(20000)
	b := make([]byte, l)

	// Windows is an utter fucking trash fire of an operating system.
	if err := windows.GetAdaptersAddresses(
		windows.AF_UNSPEC,
		windows.GAA_FLAG_INCLUDE_PREFIX, 0,
		(*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0])), &l); err != nil {
		return nil, err
	}

	var addresses []*windows.IpAdapterAddresses
	for addr := (*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0])); addr != nil; addr = addr.Next {
		addresses = append(addresses, addr)
	}

	resolvers := map[string]bool{}
	for _, addr := range addresses {
		for next := addr.FirstUnicastAddress; next != nil; next = next.Next {
			if addr.OperStatus != windows.IfOperStatusUp {
				continue
			}
			if next.Address.IP() != nil {
				for dnsServer := addr.FirstDnsServerAddress; dnsServer != nil; dnsServer = dnsServer.Next {
					ip := dnsServer.Address.IP()
					if ip.IsMulticast() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() {
						continue
					}
					if ip.To16() != nil && strings.HasPrefix(ip.To16().String(), "fec0:") {
						continue
					}
					resolvers[ip.String()] = true
				}
				break
			}
		}
	}

	// Take unique values only
	servers := []string{}
	for server := range resolvers {
		servers = append(servers, server)
	}

	// TODO: Make configurable, based on defaults in https://github.com/miekg/dns/blob/master/clientconfig.go
	return &dns.ClientConfig{
		Servers:  servers,
		Search:   []string{},
		Port:     "53",
		Ndots:    1,
		Timeout:  5,
		Attempts: 1,
	}, nil
}
