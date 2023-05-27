---
# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
---

# Features

The cunīcu daemon supports many features which are implemented by separate software modules/packages.
This structure promotes the [separation of concerns](https://en.wikipedia.org/wiki/Separation_of_concerns) within the code-base and allows for use-cases in which only subsets of features are used.
E.g. we can use cunīcu for the post-quantum safe exchange of pre-shared keys without any of the other features like peer or endpoint discovery. With very few exceptions all of the features listed below can be used separately.

Currently, the following features are implemented as separate modules:

-   [Auto-configuration of missing interface settings and link-local IP addresses](./autocfg.md) (`autocfg`)
-   [Config Synchronization](./cfgsync.md) (`cfgsync`)
-   [Peer Discovery](./pdisc.md) (`pdisc`)
-   [Endpoint Discovery](./epdisc.md) (`epdisc`)
-   [Hooks](./hooks.md) (`hooks`)
-   [Hosts-file Synchronization](./hsync.md) (`hsync`)
-   [Pre-shared Key Establishment](./pske.md) (`pske`)
-   [Route Synchronization](./rtsync.md) (`rtsync`)
