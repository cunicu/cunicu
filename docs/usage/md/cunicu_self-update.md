## cunicu self-update

Update the cunicu binary

### Synopsis


The command "self-update" downloads the latest stable release of cunicu from
GitHub and replaces the currently running binary. After download, the
authenticity of the binary is verified using the GPG signature on the release
files.


```
cunicu self-update [flags]
```

### Options

```
  -h, --help              help for self-update
  -o, --output filename   Save the downloaded file as filename (default "cunicu")
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

