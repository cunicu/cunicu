// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package nodes

import "net/url"

type SignalingNode interface {
	Node

	URL() url.URL
}
