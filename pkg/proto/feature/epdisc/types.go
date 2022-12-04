package epdisc

func NewProtocol(rp string) RelayProtocol {
	switch rp {
	case "udp", "UDP":
		return RelayProtocol_RELAY_PROTOCOL_UDP
	case "tcp":
		return RelayProtocol_RELAY_PROTOCOL_TCP
	case "dtls":
		return RelayProtocol_RELAY_PROTOCOL_DTLS
	case "tls":
		return RelayProtocol_RELAY_PROTOCOL_TLS
	}

	return RelayProtocol_RELAY_PROTOCOL_UNSPECIFIED
}

func (p RelayProtocol) ToString() string {
	switch p {
	case RelayProtocol_RELAY_PROTOCOL_UDP:
		return "udp"
	case RelayProtocol_RELAY_PROTOCOL_TCP:
		return "tcp"
	case RelayProtocol_RELAY_PROTOCOL_DTLS:
		return "dtls"
	case RelayProtocol_RELAY_PROTOCOL_TLS:
		return "tls"
	case RelayProtocol_RELAY_PROTOCOL_UNSPECIFIED:
		return "unspecified"
	}

	return "unknown"
}
