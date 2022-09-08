## cunicu status

Show current status of the cunicu daemon, its interfaces and peers

```
cunicu status [flags] [intf [peer]]
```

### Options

```
  -f, --format format       Output format (one of: human, json) (default "human")
  -h, --help                help for status
  -i, --indent              Format and indent JSON ouput (default true)
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

* [cunicu](cunicu.md)	 - cunicu

