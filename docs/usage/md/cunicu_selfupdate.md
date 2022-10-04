---
title: cunicu selfupdate
sidebar_label: selfupdate
sidebar_class_name: command-name
slug: /usage/man/selfupdate
hide_title: true
keywords:
    - manpage
---

## cunicu selfupdate

Update the cunīcu binary

### Synopsis

Downloads the latest stable release of cunīcu from GitHub and replaces the currently running binary.
After download, the authenticity of the binary is verified using the GPG signature on the release files.

```
cunicu selfupdate [flags]
```

### Options

```
  -h, --help              help for selfupdate
  -o, --output filename   Save the downloaded file as filename (default "cunicu")
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

