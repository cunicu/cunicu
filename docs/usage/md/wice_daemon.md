## wice daemon

Start the daemon

```
wice daemon [interfaces...] [flags]
```

### Options

```
  -b, --backend strings                    backend types / URLs
  -x, --community string                   Community passphrase for discovering other peers
  -c, --config strings                     Path of configuration files
  -A, --config-domain string               Perform auto-configuration via DNS
  -h, --help                               help for daemon
      --ice-candidate-type strings         usable candidate types (select from "host", "srflx", "prflx", "relay")
      --ice-check-interval duration        interval at which the agent performs candidate checks in the connecting phase
      --ice-disconnected-timout duration   time till an Agent transitions disconnected
      --ice-failed-timeout duration        time until an Agent transitions to failed after disconnected
  -k, --ice-insecure-skip-verify           skip verification of TLS certificates for secure STUN/TURN servers
      --ice-interface-filter string        regex for filtering local interfaces for ICE candidate gathering (e.g. "eth[0-9]+") (default ".*")
      --ice-keepalive-interval duration    interval netween STUN keepalives
  -L, --ice-lite                           lite agents do not perform connectivity check and only provide host candidates
      --ice-max-binding-requests uint16    maximum number of binding request before considering a pair failed
  -m, --ice-mdns                           enable local Multicast DNS discovery
      --ice-nat-1to1-ip strings            IP addresses which will be added as local server reflexive candidates
      --ice-network-type strings           usable network types (select from "udp4", "udp6", "tcp4", "tcp6")
  -P, --ice-pass string                    password for STUN/TURN credentials
      --ice-port-max uint16                maximum port for allocation policy (range: 0-65535)
      --ice-port-min uint16                minimum port for allocation policy (range: 0-65535)
      --ice-restart-timeout duration       time to wait before ICE restart
  -U, --ice-user string                    username for STUN/TURN credentials
  -f, --interface-filter string            regex for filtering Wireguard interfaces (e.g. "wg-.*") (default ".*")
  -s, --socket string                      Unix control and monitoring socket
      --socket-wait                        wait until first client connected to control socket before continuing start
  -a, --url strings                        STUN and/or TURN server addresses
  -i, --watch-interval duration            interval at which we are polling the kernel for updates on the Wireguard interfaces
  -w, --wg-config-path string              base path to search for Wireguard configuration files
  -S, --wg-config-sync                     sync Wireguard interface with configuration file (see "wg synconf")
  -u, --wg-userspace                       start userspace Wireguard daemon
```

### Options inherited from parent commands

```
  -l, --log-file string    path of a file to write logs to
  -d, --log-level string   log level (one of "debug", "info", "warn", "error", "dpanic", "panic", and "fatal") (default "info")
```

### SEE ALSO

* [wice](wice.md)	 - ɯice

