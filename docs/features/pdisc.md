---
# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
---

# Peer Discovery

The peer discovery feature can be used to find other peers.
A set of peers is identified by a common _community passphrase_.

Peers belonging to the same community will be added as WireGuard peers to the interface configuration.

No other tasks are performed.
Paired with other features like the [endpoint discovery](./epdisc.md), [auto configuration](./autocfg.md) or [route synchronization](./rtsync.md), the peer discovery is a cornerstone of a zero-configuration peer-to-peer VPN.

In addition to community passphrase, peers can be accepted by white- and blacklist filtering.

## Configuration

The following settings can be used in the main section of the [configuration file](../config/) or with-in the `interfaces` section to customize settings of an individual interface.

import ApiSchema from '@theme/ApiSchema';

<ApiSchema pointer="#/components/schemas/PeerDiscSettings" />
