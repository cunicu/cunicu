---
# SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
---

# Endpoint Discovery

The endpoint discovery finds usable WireGuard endpoint addresses for remote peers using [Interactive Connectivity Establishment (ICE)](https://en.wikipedia.org/wiki/Interactive_Connectivity_Establishment).

## Configuration

The following settings can be used in the main section of the [configuration file](../config/) or with-in the `interfaces` section to customize settings of an individual interface.

import ApiSchema from '@theme/ApiSchema';

<ApiSchema pointer="#/components/schemas/EndpointDiscoverySettings" />
