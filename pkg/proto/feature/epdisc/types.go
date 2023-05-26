// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package epdisc

func NewProtocol(rp string) RelayProtocol {
	switch rp {
	case "udp", "UDP":
		return RelayProtocol_UDP
	case "tcp":
		return RelayProtocol_TCP
	case "dtls":
		return RelayProtocol_DTLS
	case "tls":
		return RelayProtocol_TLS
	}

	return -1
}

func (p RelayProtocol) ToString() string {
	switch p {
	case RelayProtocol_UDP:
		return "udp"
	case RelayProtocol_TCP:
		return "tcp"
	case RelayProtocol_DTLS:
		return "dtls"
	case RelayProtocol_TLS:
		return "tls"
	case RelayProtocol_UNSPECIFIED_RELAY_PROTOCOL:
	}

	return "unknown"
}
