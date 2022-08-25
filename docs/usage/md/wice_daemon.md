## wice daemon

Start the daemon

```
wice daemon [interfaces...] [flags]
```

### Options

```
  -b, --backend URL                 One or more URLs to signaling backends
  -x, --community passphrase        A community passphrase for discovering other peers
  -c, --config filename             One or more filenames of configuration files
  -w, --config-path directory       The directory of WireGuard wg/wg-quick configuration files
  -W, --config-watch                Watch and synchronize changes to the WireGuard configuration files
  -A, --domain domain               A DNS domain name used for DNS auto-configuration
  -P, --password password           The password for STUN/TURN credentials
  -T, --route-table string          Kernel routing table to use (default "main")
  -s, --rpc-socket path             The path of the unix socket used by other ɯice commands
  -a, --url URL                     One or more URLs of STUN and/or TURN servers
  -U, --username username           The username for STUN/TURN credentials
  -f, --wg-interface-filter regex   A regex for filtering WireGuard interfaces (e.g. "wg-.*") (default ".*")
  -u, --wg-userspace                Create new interfaces with userspace WireGuard implementation
  -h, --help                        help for daemon
```

### Options inherited from parent commands

```
  -l, --log-file string    path of a file to write logs to
  -d, --log-level string   log level (one of: debug, info, warn, error, dpanic, panic, and fatal) (default "info")
```

### SEE ALSO

* [wice](wice.md)	 - ɯice

