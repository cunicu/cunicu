---
title: cunicu wg showconf
sidebar_label: wg showconf
sidebar_class_name: command-name
slug: /usage/man/wg/showconf
hide_title: true
keywords:
    - manpage
---

## cunicu wg showconf

Shows the current configuration and information of the provided WireGuard interface

### Synopsis

Shows the current configuration of `interface-name` in the wg(8) format.

```
cunicu wg showconf interface-name [flags]
```

### Options

```
  -h, --help                help for showconf
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

* [cunicu wg](cunicu_wg.md)	 - WireGuard commands

