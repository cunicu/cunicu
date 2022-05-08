# Configuration

This page describes the ways of configuring the ɯice daemon (`wice daemon`).

## Command Line Flags

The `wice daemon` can almost fully be configured by passing command line arguments.
A full overview is available in its [manpage](./usage/md/wice_daemon.md).

## Configuration File

Alternatively a configuration file can be used for a persistent configuration:

```yaml title="wice.yaml"
domain: 0l.de
watch_interval: 1s
community: "some-common-password"

backends:
- grpc://localhost:8080?insecure=true
- k8s:///path/to/your/kubeconfig.yaml?namespace=default
- p2p:?mdns=true&dht=true&private=false&listen-address=/ip4/0.0.0.0/tcp/1234&bootstrap-peer=/dnsaddr/bootstrap.libp2p.io/ipfs/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN&private-key=6MNeaexWoGcSKlpvJopeL0G39dqc6zrZUaZ3mbTEl1k=

# Wireguard settings
wg:
  # Use wg / wg-quick configuration files
  config:
      path: /etc/wireguard
      sync: false
    
  # Create Wireguard interfaces using bundled wireguard-go Userspace implementation
  # This will be the default if there is no Wireguard kernel module present.
  userspace: false

  # Ignore Wireguard interface which do not match this regular expression
  interface_filter: .*

  # A list of Wireguard interfaces which should be configured
  interfaces:
  - wg-vpn

# Control socket settings
socket:
  path: /var/run/wice.sock

  # Start of wice daemon will block until its unblocked via the control socket
  # Mostly useful for testing automation
  wait: false

# Interactive Connectivity Establishment
ice:
  # A list of STUN and TURN servers used by ICE
  urls:
  - stun:l.google.com:19302

  # Credentils for STUN/TURN servers configured above
  username: ""
  password: ""

  # Allow connections to STUNS/TURNS servers for which
  # we cant validate their TLS certificates
  insecure_skip_verify: false

  # Limit available network and candidate types
  network-types: [udp4, udp6, tcp4, tcp6]
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
  # This is optional. Leave them 0 for the default UDP port allocation strategy.
  port:
      max: 0
      min: 0

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

# Settings for forwarding / proxying encapsulated Wireguard traffic between
# pion/ice and the Kernel Wireguard interfaces
proxy:
  # Use NFtables to setup a port redirect / NAT for server reflexive candidates
  nft: true

  # Use a RAW socket with an attached eBPF socket filter to receive STUN packets while
  # all other data is directly received by the ListenPort of a kernel-space Wireguard interface.
  ebpf: true
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
_stun._udp.example.com.  3600 IN SRV 10 0 3478 stun.example.com.
_stuns._tcp.example.com. 3600 IN SRV 10 0 3478 stun.example.com.
_turn._udp.example.com.  3600 IN SRV 10 0 3478 turn.example.com.
_turn._tcp.example.com.  3600 IN SRV 10 0 3478 turn.example.com.
_turns._tcp.example.com. 3600 IN SRV 10 0 5349 turn.example.com.

example.com.             3600 IN TXT "wice-backend=p2p"
example.com.             3600 IN TXT "wice-community=my-community-password"
example.com.             3600 IN TXT "wice-ice-username=user1"
example.com.             3600 IN TXT "wice-ice-password=pass1"
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