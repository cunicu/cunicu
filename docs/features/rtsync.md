---
# SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
---

# Route Synchronization

The route synchronization feature keeps the kernel routing table in sync with WireGuard's _AllowedIPs_ setting.

This synchronization is bi-directional:
-   Networks with are found in a Peers AllowedIP list will be installed as a kernel route.
-   Kernel routes with the peers link-local IP address as next-hop will be added to the Peers _AllowedIPs_ list.

This rather simple feature allows user to pair cunicu with a software routing daemon like [Bird2](https://bird.network.cz/) while using a single WireGuard interface with multiple peer-to-peer links.

## Configuration

The following settings can be used in the main section of the [configuration file](../config/) or with-in the `interfaces` section to customize settings of an individual interface.

import ApiSchema from '@theme/ApiSchema';

<ApiSchema pointer="#/components/schemas/RouteSyncSettings" />
