// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package nodes

import "net/url"

type SignalingNode interface {
	Node

	URL() *url.URL
}
