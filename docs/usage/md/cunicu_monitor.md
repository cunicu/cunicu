---
title: cunicu monitor
sidebar_label: monitor
sidebar_class_name: command-name
slug: /usage/man/monitor
hide_title: true
keywords:
    - manpage
---

## cunicu monitor

Monitor the cunīcu daemon for events

```
cunicu monitor [flags]
```

### Options

```
  -f, --format format       Output format (one of: json, logger, human) (default "human")
  -h, --help                help for monitor
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

