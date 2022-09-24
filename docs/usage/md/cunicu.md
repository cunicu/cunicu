---
title: cunicu
sidebar_class_name: command-name
slug: /usage/man/
hide_title: true
keywords:
    - manpage
---

## cunicu

cunīcu is a user-space daemon managing WireGuard® interfaces to establish peer-to-peer connections in harsh network environments.

### Synopsis

It relies on the awesome pion/ice package for the interactive connectivity establishment as well as bundles the Go user-space implementation of WireGuard in a single binary for environments in which WireGuard kernel support has not landed yet.

### Options

```
  -C, --color string       Enable colorization of output (one of: auto, always, never) (default "auto")
  -l, --log-file string    path of a file to write logs to
  -d, --log-level string   log level (one of: debug, info, warn, error, dpanic, panic, and fatal) (default "info")
  -v, --verbose int        verbosity level
  -h, --help               help for cunicu
```

### SEE ALSO

* [cunicu addresses](cunicu_addresses.md)	 - Calculate link-local IPv4 and IPv6 addresses from a WireGuard X25519 public key
* [cunicu completion](cunicu_completion.md)	 - Generate the autocompletion script for the specified shell
* [cunicu config](cunicu_config.md)	 - Manage configuration of a running cunīcu daemon.
* [cunicu daemon](cunicu_daemon.md)	 - Start the daemon
* [cunicu monitor](cunicu_monitor.md)	 - Monitor the cunīcu daemon for events
* [cunicu relay](cunicu_relay.md)	 - Start relay API server
* [cunicu reload](cunicu_reload.md)	 - Reload the configuration of the cunīcu daemon
* [cunicu restart](cunicu_restart.md)	 - Restart the cunīcu daemon
* [cunicu selfupdate](cunicu_selfupdate.md)	 - Update the cunīcu binary
* [cunicu signal](cunicu_signal.md)	 - Start gRPC signaling server
* [cunicu status](cunicu_status.md)	 - Show current status of the cunīcu daemon, its interfaces and peers
* [cunicu stop](cunicu_stop.md)	 - Shutdown the cunīcu daemon
* [cunicu sync](cunicu_sync.md)	 - Synchronize cunīcu daemon state
* [cunicu version](cunicu_version.md)	 - Show version of the cunīcu binary and optionally also a running daemon
* [cunicu wg](cunicu_wg.md)	 - WireGuard commands

