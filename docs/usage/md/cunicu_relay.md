---
title: cunicu relay
sidebar_label: relay
sidebar_class_name: command-name
slug: /usage/man/relay
hide_title: true
keywords:
    - manpage
---

## cunicu relay

Start relay API server

### Synopsis

This command starts a gRPC server providing cunicu agents with a list of available STUN and TURN servers.

**Note:** Currently this command does not run a TURN server itself. But relies on an external server like Coturn.

With this feature you can distribute a list of available STUN/TURN servers easily to a fleet of agents.
It also allows to issue short-lived HMAC-SHA1 credentials based the proposed TURN REST API and thereby static long term credentials.

The command expects a list of STUN or TURN URLs according to RFC7065/RFC7064 with a few extensions:

- A secret for the TURN REST API can be provided by the 'secret' query parameter
  - Example: turn:server.com?secret=rest-api-secret

- A time-to-live to the TURN REST API secrets can be provided by the 'ttl' query parameter
  - Example: turn:server.com?ttl=1h

- Static TURN credentials can be provided by the URIs user info
  - Example: turn:user1:pass1@server.com


```
cunicu relay URL... [flags]
```

### Examples

```
relay turn:server.com?secret=rest-api-secret&ttl=1h
```

### Options

```
  -h, --help            help for relay
  -L, --listen string   listen address (default ":8080")
  -S, --secure          listen with TLS
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

