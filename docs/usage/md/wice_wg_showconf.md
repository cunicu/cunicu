## wice wg showconf

Shows the current configuration and device information

### Synopsis

Sets the current configuration of <interface> to the contents of <configuration-filename>, which must be in the wg(8) format.

```
wice wg showconf [flags] <interface>
```

### Options

```
  -h, --help                help for showconf
  -s, --rpc-socket string   Unix control and monitoring socket (default "/var/run/wice.sock")
```

### Options inherited from parent commands

```
  -C, --color string       Enable colorization of output (one of: auto, always, never) (default "auto")
  -l, --log-file string    path of a file to write logs to
  -d, --log-level string   log level (one of: debug, info, warn, error, dpanic, panic, and fatal) (default "info")
```

### SEE ALSO

* [wice wg](wice_wg.md)	 - WireGuard commands

