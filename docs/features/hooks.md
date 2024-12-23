---
# SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
---

# Hooks

The hooks feature allows the user to configure a list of hook functions which are triggered by certain events within the daemon.

## Configuration

The following settings can be used in the main section of the [configuration file](../config/) or with-in the `interfaces` section to customize settings of an individual interface.

import ApiSchema from '@theme/ApiSchema';

<ApiSchema pointer="#/components/schemas/HooksSettings"/>
