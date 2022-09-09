---
title: cunicu version
sidebar_label: version
sidebar_class_name: command-name
slug: /usage/man/version
hide_title: true
keywords:
    - manpage
---

## cunicu version

Show version of the cunīcu binary and optionally also a running daemon

```
cunicu version [flags]
```

### Examples

```
$ sudo cunicu version
client: v0.1.2 (os=linux, arch=arm64, commit=b22ee3e7, branch=master, built-at=2022-09-09T13:44:22+02:00, built-by=goreleaser)
daemon: v0.1.2 (os=linux, arch=arm64, commit=b22ee3e7, branch=master, built-at=2022-09-09T13:44:22+02:00, built-by=goreleaser)
```

### Options

```
  -f, --format format   Output format (one of: human, json) (default "human")
  -h, --help            help for version
  -s, --short           Only show version and nothing else
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

