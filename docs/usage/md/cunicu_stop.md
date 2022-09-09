---
title: cunicu stop
sidebar_label: stop
sidebar_class_name: command-name
slug: /usage/man/stop
hide_title: true
keywords:
    - manpage
---

## cunicu stop

Shutdown the cunīcu daemon

```
cunicu stop [flags]
```

### Options

```
  -h, --help                help for stop
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

