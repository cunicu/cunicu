---
title: cunicu sync
sidebar_label: sync
sidebar_class_name: command-name
slug: /usage/man/sync
hide_title: true
keywords:
    - manpage
---

## cunicu sync

Synchronize cunīcu daemon state

### Synopsis

Synchronizes the internal daemon state with kernel routes, interfaces and addresses

```
cunicu sync [flags]
```

### Options

```
  -h, --help                help for sync
  -s, --rpc-socket string   Unix control and monitoring socket (default "/var/run/cunicu.sock")
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

