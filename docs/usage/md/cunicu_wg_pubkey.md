---
title: cunicu wg pubkey
sidebar_label: wg pubkey
sidebar_class_name: command-name
slug: /usage/man/wg/pubkey
hide_title: true
keywords:
    - manpage
---

## cunicu wg pubkey

Calculates a public key and prints it in base64 to standard output.

### Synopsis

Calculates a public key and prints it in base64 to standard output from a corresponding private key (generated with genkey) given in base64 on standard input.

```
cunicu wg pubkey [flags]
```

### Examples

```
# A private key and a corresponding public key may be generated at once by calling:
$ umask 077
$ wg genkey | tee private.key | wg pubkey > public.key
```

### Options

```
  -h, --help   help for pubkey
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

