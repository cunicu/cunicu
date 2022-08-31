package nodes

import "net/url"

type SignalingNode interface {
	Node

	URL() *url.URL
}
