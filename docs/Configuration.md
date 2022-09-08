# Configuration

This page describes the ways of configuring the cunicu daemon (`wice daemon`).

## Command Line Flags

The `wice daemon` can almost fully be configured by passing command line arguments.
A full overview is available in its [manpage](./usage/md/wice_daemon.md).

## Configuration File

Alternatively a configuration file can be used for a persistent configuration:

```yaml title="wice.yaml"
watch_interval: 1s

backends:
- grpc://localhost:8080?insecure=true&skip_verify=true
- k8s:///path/to/your/kubeconfig.yaml?namespace=default

# WireGuard settings
wireguard:  
  # Create WireGuard interfaces using bundled wireguard-go Userspace implementation
  # This will be the default if there is no WireGuard kernel module present.
  userspace: false

  # Ignore WireGuard interface which do not match this regular expression
  interface_filter: .*

  # A list of WireGuard interfaces which should be configured
  interfaces:
  - wg-vpn

  # Port range for ListenPort setting of newly created WireGuard interfaces
  # wice will select the first available port in this range.
  port:
    min: 52820
    max: 65535

# Control socket settings
socket:
  path: /var/run/wice.sock

  # Start of wice daemon will block until its unblocked via the control socket
  # Mostly useful for testing automation
  wait: false

# Synchronize WireGuard interface configurations with wg(8) config-files.
config_sync:
  enabled: false
  
  # Directory where Wireguard configuration files are located.
  # We expect the same format as used by wg(8) and wg-quick(8).
  # Filenames must match the interface name with a '.conf' suffix.
  path: /etc/wireguard

  # Watch the configuration files for changes and apply them accordingly.
  watch: false
  
# Synchronize WireGuard AllowedIPs with Kernel routing table
route_sync:
  enabled: true

  table: main

# Discovery of other WireGuard peers
peer_disc:
  enabled: true

  # A list of WireGuard public keys which are accepted peers
  whitelist:
  - coNsGPwVPdpahc8U+dbbWGzTAdCd6+1BvPIYg10wDCI=
  - AOZzBaNsoV7P8vo0D5UmuIJUQ7AjMbHbGt2EA8eAuEc=

  # A passphrase shared among all peers of the same community
  community: "some-common-password"

# Discovery of WireGuard endpoint addressesendpoint_disc:
  enabled: true

  # Interactive Connectivity Establishment parameters
  ice:
    # A list of STUN and TURN servers used by ICE
    urls:
    - stun:stun.l.google.com:19302

    # Credentials for STUN/TURN servers configured above
    username: ""
    password: ""

    # Allow connections to STUNS/TURNS servers for which
    # we cant validate their TLS certificates
    insecure_skip_verify: false

    # Limit available network and candidate types
    network_types: [udp4, udp6, tcp4, tcp6]
    candidate_types: [host, srflx, prflx ,relay]

    # Regular expression whitelist of interfaces which are used to gather ICE candidates.
    interface_filter: .*

    # Lite agents do not perform connectivity check and only provide host candidates.
    lite: false

    # Attempt to find candidates via mDNS discovery
    mdns: false

    # Sets the max amount of binding requests the agent will send over a candidate pair for validation or nomination.
    # If after the the configured number, the candidate is yet to answer a binding request or a nomination we set the pair as failed.
    max_binding_requests: 7

    # SetNAT1To1IPs sets a list of external IP addresses of 1:1 (D)NAT and a candidate type for which the external IP address is used.
    # This is useful when you are host a server using Pion on an AWS EC2 instance which has a private address, behind a 1:1 DNAT with a public IP (e.g. Elastic IP).
    # In this case, you can give the public IP address so that Pion will use the public IP address in its candidate instead of the private IP address.
    nat_1to1_ips: []

    # Limit the port range used by ICE
    port:
        min: 49152
        max: 65535

    # The check interval controls how often our task loop runs when in the connecting state.
    check_interval: 200ms
    
    # If the duration is 0, the ICE Agent will never go to disconnected
    disconnected_timeout: 5s

    # If the duration is 0, we will never go to failed.
    failed_timeout: 5s
    restart_timeout: 5s

    # Determines how often should we send ICE keepalives (should be less then connection timeout above).
    # A keepalive interval of 0 means we never send keepalive packets
    keepalive_interval: 2s
```

## Environment Variables

All the settings from the configuration file can also be passed via environment variables by following the following rules:

-   Convert the setting name to uppercase
-   Prefixing the setting name with `WICE_`
-   Nested settings are separated by underscores

**Example:** The setting `endpoint_disc.ice.max_binding_requests` can be set by the environment variable `WICE_ENDPOINT_DISC_ICE_MAX_BINDING_REQUESTS`

**Note:** Setting lists such as `endpoint_disc.ice.urls` or `backends` can currently not be set via environment variables.

## DNS Auto-configuration

cunicu als supports retrieving parts of the configuration via DNS lookups.

When `wice daemon` is started with a `--domain example.com` parameter it will look for the following DNS records to obtain its configuration.

STUN and TURN servers used for ICE are retrieved by SVR lookups and other cunicu settings are retrieved via TXT lookups: 

```text
_stun._udp.example.com.  3600 IN SRV 10 0 3478 stun.example.com.
_stuns._tcp.example.com. 3600 IN SRV 10 0 3478 stun.example.com.
_turn._udp.example.com.  3600 IN SRV 10 0 3478 turn.example.com.
_turn._tcp.example.com.  3600 IN SRV 10 0 3478 turn.example.com.
_turns._tcp.example.com. 3600 IN SRV 10 0 5349 turn.example.com.

example.com.             3600 IN TXT "wice-backend=p2p"
example.com.             3600 IN TXT "wice-peer-disc-community=my-community-password"
example.com.             3600 IN TXT "wice-endpoint-disc-ice-username=user1"
example.com.             3600 IN TXT "wice-endpoint-disc-ice-password=pass1"
example.com.             3600 IN TXT "wice-config=https://example.com/wice.yaml"
```

**Note:** The `wice-backend` and `wice-config` TXT records can be provided multiple times. Others not.

## Remote Configuration File

When `wice daemon` can be started with `--config` options pointing to HTTPS URIs.
cunicu will download all configuration files in the order they are specified on the command line and merge them subsequently.

This feature can be combined with the DNS auto-configuration method by providing a TXT record pointing to the configuration file:

```text
example.com.             3600 IN TXT "wice-config=https://example.com/wice.yaml"
```

**Note:** Remote configuration files must be fetched via HTTPS if they are not hosted locally and required a trusted server certificate.