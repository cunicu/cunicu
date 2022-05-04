## wice self-update

Update the wice binary

### Synopsis


The command "self-update" downloads the latest stable release of wice from
GitHub and replaces the currently running binary. After download, the
authenticity of the binary is verified using the GPG signature on the release
files.


```
wice self-update [flags]
```

### Options

```
  -h, --help              help for self-update
      --output filename   Save the downloaded file as filename (default: running binary itself)
```

### Options inherited from parent commands

```
  -l, --log-file string    path of a file to write logs to
  -d, --log-level string   log level (one of "debug", "info", "warn", "error", "dpanic", "panic", and "fatal") (default "info")
```

### SEE ALSO

* [wice](wice.md)	 - ɯice

