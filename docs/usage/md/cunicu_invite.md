---
title: cunicu invite
sidebar_label: invite
sidebar_class_name: command-name
slug: /usage/man/invite
hide_title: true
keywords:
    - manpage
---

## cunicu invite

Add a new peer to the local daemon configuration and return the required configuration for this new peer

```
cunicu invite [interface] [flags]
```

### Options

```
  -h, --help                help for invite
  -L, --listen-port int     Listen port for generated config (default 51820)
  -Q, --qr-code             Show config as QR code in terminal
  -s, --rpc-socket string   Unix control and monitoring socket (default "/var/run/cunicu.sock")
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

