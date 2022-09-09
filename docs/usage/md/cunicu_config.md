---
title: cunicu config
sidebar_label: config
sidebar_class_name: command-name
slug: /usage/man/config
hide_title: true
keywords:
    - manpage
---

## cunicu config

Manage configuration of a running cunīcu daemon.

### Synopsis




### Options

```
  -h, --help                help for config
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
* [cunicu config get](cunicu_config_get.md)	 - Get current value of a configuration setting
* [cunicu config set](cunicu_config_set.md)	 - Update the value of a configuration setting

