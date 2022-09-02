## wice completion zsh

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(wice completion zsh); compdef _wice wice

To load completions for every new session, execute once:

#### Linux:

	wice completion zsh > "${fpath[1]}/_wice"

#### macOS:

	wice completion zsh > $(brew --prefix)/share/zsh/site-functions/_wice

You will need to start a new shell for this setup to take effect.


```
wice completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
  -C, --color string       Enable colorization of output (one of: auto, always, never) (default "auto")
  -l, --log-file string    path of a file to write logs to
  -d, --log-level string   log level (one of: debug, info, warn, error, dpanic, panic, and fatal) (default "info")
```

### SEE ALSO

* [wice completion](wice_completion.md)	 - Generate the autocompletion script for the specified shell

