# Configuration

This page describes the ways of configuring the ɯice daemon (`wice daemon`).

## Command Line Flags

The `wice daemon` can almost fully be configured by passing command line arguments.
A full overview is available in its [manpage](./usage/md/wice_daemon.md).

## Configuration File

Alternatively a configuration file can be used for a persistent configuration:

```yaml
backends:
- p2p

domain: example.com

community: "some-common-password"

watch_interval: 2s

ice:
    url:
    - stun:l.google.com:19302

    username: ""
    password: ""

    network-types: [udp4, udp6, tcp4, tcp6]
    candidate_types: [host, srflx, prflx ,relay]

    insecure_skip_verify: false
    interface_filter: .*
    lite: false
    mdns: false

    max_binding_requests: 7
    nat_1to1_ips: []

    port:
        max: 0
        min: 0

    check_interval: 200ms
    disconnected_timeout: 5s
    failed_timeout: 5s
    restart_timeout: 5s
    keepalive_interval: 2s

wg:
    config:
        path: /etc/wireguard
        sync: false
    interface_filter: .*
    userspace: false
```

## Environment Variables

All the settings from the configuration file can also be passed via environment variables by following the following rules:

- Convert the setting name to uppercase
- Prefixing the setting name with `WICE_`
- Nested settings are separated by underscores

**Example:** The setting `ice.max_binding_requests` can be set by the environment variable `WICE_ICE_MAX_BINDING_REQUESTS`

**Note:** Setting lists such as `ice.urls` or `backends` can currently not be set via environment variables.

## DNS Auto-configuration

ɯice als supports retrieving parts of the configuration via DNS lookups.

When `wice daemon` is started with a `--config-domain example.com` parameter it will look for the following DNS records to obtain its configuration.

STUN and TURN servers used for ICE are retrieved by SVR lookups and other ɯice settings are retrieved via TXT lookups: 

```
_stun._udp.example.com.  3600 IN SRV 10 0 3478 stun.example.com
_stuns._tcp.example.com. 3600 IN SRV 10 0 3478 stun.example.com
_turn._udp.example.com.  3600 IN SRV 10 0 3478 turn.example.com
_turns._tcp.example.com. 3600 IN SRV 10 0 5349 turn.example.com

example.com.             3600 IN TXT "wice-backend=p2p"
example.com.             3600 IN TXT "wice-community=my-community-password"
example.com.             3600 IN TXT "wice-ice-username=user1"
example.com.             3600 IN TXT "wice-ice-passpassword=pass1"
example.com.             3600 IN TXT "wice-config=https://example.com/wice.yaml"
```

**Note:** The `wice-backend` and `wice-config` TXT records can be provided multiple times. Others not.

## Remote Configuration File

When `wice daemon` can be started with `--config` options pointing to HTTPS URIs.
ɯice will download all configuration files in the order they are specified on the command line and merge them subsequently.

This feature can be combined with the DNS auto-configuration method by providing a TXT record pointing to the configuration file:

```
example.com.             3600 IN TXT "wice-config=https://example.com/wice.yaml"
```

**Note:** Remote configuration files must be fetched via HTTPS if they are not hosted locally and required a trusted server certificate.