---
title: cunicu self-update
sidebar_label: self-update
sidebar_class_name: command-name
slug: /usage/man/self-update
hide_title: true
keywords:
    - manpage
---

## cunicu self-update

Update the cunīcu binary

### Synopsis

The command "self-update" downloads the latest stable release of cunicu from
GitHub and replaces the currently running binary. After download, the
authenticity of the binary is verified using the GPG signature on the release
files.


```
cunicu self-update [flags]
```

### Options

```
  -h, --help              help for self-update
  -o, --output filename   Save the downloaded file as filename (default "cunicu")
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

