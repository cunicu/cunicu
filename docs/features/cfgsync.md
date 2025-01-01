---
# SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
---

# Config Synchronization

The config synchronization feature keeps interface configuration provided via configuration files in sync with the kernel.

## Configuration

The following settings can be used in the main section of the [configuration file](../config/) or with-in the `interfaces` section to customize settings of an individual interface.

import ApiSchema from '@theme/ApiSchema';

<ApiSchema pointer="#/components/schemas/ConfigSyncSettings" />
