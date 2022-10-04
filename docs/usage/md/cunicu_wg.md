---
title: cunicu wg
sidebar_label: wg
sidebar_class_name: command-name
slug: /usage/man/wg
hide_title: true
keywords:
    - manpage
---

## cunicu wg

WireGuard commands

### Synopsis

The wg sub-command mimics the wg(8) commands of the wireguard-tools package.
In contrast to the wg(8) command, the cunico sub-command delegates it tasks to a running cunucu daemon.

Currently, only a subset of the wg(8) are supported.

### Options

```
  -h, --help   help for wg
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
* [cunicu wg genkey](cunicu_wg_genkey.md)	 - Generates a random private key in base64 and prints it to standard output.
* [cunicu wg genpsk](cunicu_wg_genpsk.md)	 - Generates a random preshared key in base64 and prints it to standard output.
* [cunicu wg pubkey](cunicu_wg_pubkey.md)	 - Calculates a public key and prints it in base64 to standard output.
* [cunicu wg show](cunicu_wg_show.md)	 - Shows current WireGuard configuration and runtime information of specified [interface].
* [cunicu wg showconf](cunicu_wg_showconf.md)	 - Shows the current configuration and information of the provided WireGuard interface

