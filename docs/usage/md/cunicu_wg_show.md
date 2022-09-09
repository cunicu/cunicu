## cunicu wg show

Shows current WireGuard configuration and runtime information of specified <interface>.

### Synopsis

Shows current WireGuard configuration and runtime information of specified <interface>.
		
If no <interface> is specified, <interface> defaults to all.

If 'interfaces' is specified, prints a list of all WireGuard interaces, one per line, and quits.

If no options are given after the interface specification, then prints a list of all attributes in a visually pleasing way meant for the terminal.
Otherwise, prints specified information grouped by newlines and tabs, meant to be used in scripts.

For this script-friendly display, if 'all' is specified, then the first field for all categories of information is the interface name.

If 'dump' is specified, then several lines are printed; the first contains in order separated by tab: private-key, public-key, listen-port, fwmark.
Subsequent lines are printed for each peer and contain in order separated by tab: public-key, preshared-key, endpoint, allowed-ips, latest-handshake, transfer-rx, transfer-tx, persistent-keepalive.

```
cunicu wg show [flags] { <interface> | all | interfaces } [public-key | private-key | listen-port | fwmark | peers | preshared-keys | endpoints | allowed-ips | latest-handshakes | transfer | persistent-keepalive | dump]
```

### Options

```
  -h, --help                help for show
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

