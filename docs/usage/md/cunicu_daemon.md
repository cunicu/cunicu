---
title: cunicu daemon
sidebar_label: daemon
sidebar_class_name: command-name
slug: /usage/man/daemon
hide_title: true
keywords:
    - manpage
---

## cunicu daemon

Start the daemon

```
cunicu daemon [interface-names...] [flags]
```

### Examples

```
$ cunicu daemon -u -x mysecretpass wg0
```

### Options

```
  -b, --backend URL                         One or more URLs to signaling backends
  -x, --community passphrase                A passphrase shared with other peers in the same community
  -c, --config filename                     One or more filenames of configuration files
  -E, --discover-endpoints                  Enable ICE endpoint discovery (default true)
  -P, --discover-peers                      Enable peer discovery (default true)
  -D, --domain domain                       A DNS domain name used for DNS auto-configuration
      --ice-candidate-type candidate-type   Usable candidate-types (one of host, srflx, prflx, relay)
      --ice-network-type network-type       Usable network-types (one of udp4, udp6, tcp4, tcp6)
  -p, --password password                   The password for STUN/TURN credentials
  -T, --routing-table int                   Kernel routing table to use (default 254)
  -s, --rpc-socket path                     The path of the unix socket used by other cunicu commands
      --rpc-wait                            Wait until first client connected to control socket before continuing start
  -C, --sync-config                         Enable synchronization of configuration files (default true)
  -H, --sync-hosts                          Enable synchronization of /etc/hosts file (default true)
  -R, --sync-routes                         Enable synchronization of AllowedIPs with Kernel routes (default true)
  -a, --url URL                             One or more URLs of STUN and/or TURN servers
  -u, --username username                   The username for STUN/TURN credentials
  -w, --watch-config                        Watch configuration files for changes and apply changes at runtime.
  -i, --watch-interval duration             An interval at which we are periodically polling the kernel for updates on WireGuard interfaces
  -U, --wg-userspace                        Use user-space WireGuard implementation for newly created interfaces
  -h, --help                                help for daemon
```

### Options inherited from parent commands

```
  -q, --color string       Enable colorization of output (one of: auto, always, never) (default "auto")
  -l, --log-file string    path of a file to write logs to
  -d, --log-level string   log level (one of: debug, info, warn, error, dpanic, panic, and fatal) (default "info")
  -v, --verbose int        verbosity level
```

### SEE ALSO

* [cunicu](cunicu.md)	 - cunīcu is a user-space daemon managing WireGuard® interfaces to establish peer-to-peer connections in harsh network environments.

