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

Calculate link-local IPv4 and IPv6 addresses from a WireGuard X25519 public key

### Synopsis

cunīcu auto-configuration feature derives and assigns link-local IPv4 and IPv6 addresses based on the public key of the WireGuard interface.
This sub-command accepts a WireGuard public key on the standard input and prints out the calculated IP addresses on the standard output.


```
cunicu addresses [flags]
```

### Examples

```
$ wg genkey | wg pubkey | cunicu addresses
fe80::e3be:9673:5a98:9348/64
169.254.29.188/16
```

### Options

```
  -h, --help   help for addresses
  -4, --ipv4   Print IPv4 address only
  -6, --ipv6   Print IPv6 address only
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

