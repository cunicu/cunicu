---
# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
---

# Usage

cunīcu is a command line tool which is distributed as a single binary file.
Currently, no graphical interface is available.

The main `cunicu` command features several sub-commands of which `cunicu daemon` starts the cunīcu agent of which usally once instance runs on each peer.

For a detailed documentation of the `cunicu` command-line tool please have a look at the following page:

**[`cunicu`](./md/cunicu.md)**

## Example Use-cases

## Zero-configuration (almost)

**Invocation:** `cunicu daemon --community a-secret-shared-passphrase wg0`

## Start user-space WireGuard daemon

**Invocation:** `cunicu daemon --wg-userspace wg0`

## Co-exist with wg-quick, NetworkManager and or Manual WireGuard configuration

**Invocation:** `cunicu daemon`
