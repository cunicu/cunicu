## wice wg pubkey

Calculates a public key and prints it in base64 to standard output.

### Synopsis

Calculates a public key and prints it in base64 to standard output from a corresponding private key (generated with genkey) given in base64 on standard input.

A private key and a corresponding public key may be generated at once by calling:
$ umask 077
$ wg genkey | tee private.key | wg pubkey > public.key
		

```
wice wg pubkey [flags]
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
```

### SEE ALSO

* [wice wg](wice_wg.md)	 - WireGuard commands

