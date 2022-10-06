---
title: cunicu addresses
sidebar_label: addresses
sidebar_class_name: command-name
slug: /usage/man/addresses
hide_title: true
keywords:
    - manpage
---

## cunicu addresses

Derive IPv4 and IPv6 addresses from a WireGuard X25519 public key

### Synopsis

cunīcu auto-configuration feature derives and assigns IPv4 and IPv6 addresses based on the public key of the WireGuard interface.
This sub-command accepts a WireGuard public key on the standard input and prints out the calculated IP addresses on the standard output.


```
cunicu addresses [flags]
```

### Examples

```
$ wg genkey | wg pubkey | cunicu addresses
fc2f:9a4d:777f:7a97:8197:4a5d:1d1b:ed79
10.237.119.127
```

### Options

```
  -h, --help   help for addresses
  -m, --mask   Print CIDR mask
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

