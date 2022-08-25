## wice completion fish

Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	wice completion fish | source

To load completions for every new session, execute once:

	wice completion fish > ~/.config/fish/completions/wice.fish

You will need to start a new shell for this setup to take effect.


```
wice completion fish [flags]
```

### Options

```
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
  -l, --log-file string    path of a file to write logs to
  -d, --log-level string   log level (one of: debug, info, warn, error, dpanic, panic, and fatal) (default "info")
```

### SEE ALSO

* [wice completion](wice_completion.md)	 - Generate the autocompletion script for the specified shell

