## wice self-update

Update the ɯice binary

### Synopsis


The command "self-update" downloads the latest stable release of ɯice from
GitHub and replaces the currently running binary. After download, the
authenticity of the binary is verified using the GPG signature on the release
files.


```
wice self-update [flags]
```

### Options

```
  -h, --help              help for self-update
  -o, --output filename   Save the downloaded file as filename (default "/tmp/go-build2034746773/b001/exe/wice")
```

### Options inherited from parent commands

```
  -l, --log-file string    path of a file to write logs to
  -d, --log-level string   log level (one of: debug, info, warn, error, dpanic, panic, and fatal) (default "info")
```

### SEE ALSO

* [wice](wice.md)	 - ɯice

