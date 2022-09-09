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
  -A, --auto-config                         Enable setup of link-local addresses and missing interface options (default true)
  -b, --backend URL                         One or more URLs to signaling backends
  -x, --community passphrase                A passphrase shared with other peers in the same community
  -c, --config filename                     One or more filenames of configuration files
  -w, --config-path directory               The directory of WireGuard wg/wg-quick configuration files
  -S, --config-sync                         Enable synchronization of WireGuard configuration files (default true)
  -W, --config-watch                        Watch and synchronize changes to the WireGuard configuration files
  -D, --domain domain                       A DNS domain name used for DNS auto-configuration
  -I, --endpoint-disc                       Enable ICE endpoint discovery (default true)
  -H, --host-sync                           Enable synchronization of /etc/hosts file (default true)
      --ice-candidate-type candidate-type   Usable candidate-types (one of host, srflx, prflx, relay)
      --ice-network-type network-type       Usable network-types (one of udp4, udp6, tcp4, tcp6)
  -P, --password password                   The password for STUN/TURN credentials
  -R, --route-sync                          Enable synchronization of AllowedIPs and Kernel routing table (default true)
  -T, --route-table string                  Kernel routing table to use (default "main")
  -s, --rpc-socket path                     The path of the unix socket used by other cunicu commands
      --rpc-wait                            Wait until first client connected to control socket before continuing start
  -a, --url URL                             One or more URLs of STUN and/or TURN servers
  -U, --username username                   The username for STUN/TURN credentials
  -i, --watch-interval duration             An interval at which we are periodically polling the kernel for updates on WireGuard interfaces
  -f, --wg-interface-filter regex           A regex for filtering WireGuard interfaces (e.g. "wg-.*") (default ".*")
  -u, --wg-userspace                        Create new interfaces with userspace WireGuard implementation
  -h, --help                                help for daemon
```

### Options inherited from parent commands

```
  -C, --color string       Enable colorization of output (one of: auto, always, never) (default "auto")
  -l, --log-file string    path of a file to write logs to
  -d, --log-level string   log level (one of: debug, info, warn, error, dpanic, panic, and fatal) (default "info")
  -v, --verbose int        verbosity level
```

### SEE ALSO

* [cunicu](cunicu.md)	 - cunīcu is a user-space daemon managing WireGuard® interfaces to establish peer-to-peer connections in harsh network environments.

