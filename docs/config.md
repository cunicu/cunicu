---
sidebar_position: 7
# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
---

# Configuration

This page describes the ways of configuring the cunīcu daemon (`cunicu daemon`).

## Command Line Flags

Basic options of `cunicu daemon` can be configured by passing command line arguments.
A full overview is available in its [manpage](./usage/md/cunicu_daemon.md).

## Configuration File

For more advanced setups, a configuration file can be used for a persistent configuration:

Please have a look at the [example configuration file](./config-reference.md) for a full reference of all available settings.

## Environment Variables

All the settings from the configuration file can also be passed via environment variables by following the following rules:

-   Convert the setting name to uppercase
-   Prefixing the setting name with `CUNICU_`
-   Nested settings are separated by underscores

**Example:** The setting `ice.max_binding_requests` can be set by the environment variable `CUNICU_ICE_MAX_BINDING_REQUESTS`

:::note
Setting lists such as `ice.urls` or `backends` can currently not be set via environment variables.
:::

## At Runtime

cunīcu's configuration can also be updated at runtime, elevating the need to restart the daemon to avoid interruption of connectivity.

Please have a look at the [`cunicu config`](./usage/md/cunicu_config.md) commands.

## DNS Auto-configuration

cunīcu als supports retrieving parts of the configuration via DNS lookups.
This is useful for corporate environments in which a fleet of cunīcu daemon need to be configured centrally.

In this case `cunicu daemon` is started one or more `--domain example.com` parameters to look for the following DNS records to obtain its configuration.

STUN and TURN servers used for ICE are retrieved by SVR lookups and other cunīcu settings are retrieved via `SRV` and `TXT` lookups: 

```text
_stun._udp.example.com.  3600 IN SRV 10 0 3478 stun.example.com.
_stuns._tcp.example.com. 3600 IN SRV 10 0 3478 stun.example.com.
_turn._udp.example.com.  3600 IN SRV 10 0 3478 turn.example.com.
_turn._tcp.example.com.  3600 IN SRV 10 0 3478 turn.example.com.
_turns._tcp.example.com. 3600 IN SRV 10 0 5349 turn.example.com.

example.com.             3600 IN TXT "cunicu-config=https://example.com/cunicu.yaml"
example.com.             3600 IN TXT "cunicu-backend=grpc://signal.example.com:443"
example.com.             3600 IN TXT "cunicu-community=my-community-password"
example.com.             3600 IN TXT "cunicu-ice-username=user1"
example.com.             3600 IN TXT "cunicu-ice-password=pass1"
```

:::note
The `cunicu-backend` and `cunicu-config` TXT records can be provided multiple times. Others not.
:::

## Remote Configuration Files

When `cunicu daemon` can be started with `--config` options pointing to HTTPS URIs:

```bash
cunicu daemon --config http://example.com/cunicu.yaml
```

cunīcu will download all configuration files in the order they are specified on the command line and merge them subsequently.

This feature can be combined with the DNS auto-configuration method by providing a TXT record pointing to the configuration file:

```text
example.com. 3600 IN TXT "cunicu-config=https://example.com/cunicu.yaml"
```

:::note
Remote configuration files must be fetched via HTTPS if they are not hosted locally and required a trusted server certificate.
:::

## Auto-reload

cunīcu watches local and remote files as well as the DNS configuration for changes and automatically reloads its configuration from them whenever a change has been detected.

For local files the change is detected by [inotify(7)](https://man7.org/linux/man-pages/man7/inotify.7.html).
For remote sources, cunīcu periodically checks the `Last-Modified` and `Etag` headers in case of HTTP files or the DNS zone's [SOA serial number](https://en.wikipedia.org/wiki/SOA_record#Structure) to detect changes without request the full remote source.

:::note
Configuration file distributed via `conicu-config` DNS TXT record are not yet monitored for changes.
:::
