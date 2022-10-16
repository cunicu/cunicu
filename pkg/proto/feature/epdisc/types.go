package epdisc

func NewProtocol(rp string) Protocol {
	switch rp {
	case "udp", "UDP":
		return Protocol_UDP
	case "tcp":
		return Protocol_TCP
	case "dtls":
		return Protocol_DTLS
	case "tls":
		return Protocol_TLS
	}

	return Protocol_UNKNOWN_PROTOCOL
}

func (p Protocol) ToString() string {
	switch p {
	case Protocol_UDP:
		return "udp"
	case Protocol_TCP:
		return "tcp"
	case Protocol_DTLS:
		return "dtls"
	case Protocol_TLS:
		return "tls"
	}

	return "unknown"
}
