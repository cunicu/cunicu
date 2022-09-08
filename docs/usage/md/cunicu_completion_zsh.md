## cunicu completion zsh

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(cunicu completion zsh); compdef _cunicu cunicu

To load completions for every new session, execute once:

#### Linux:

	cunicu completion zsh > "${fpath[1]}/_cunicu"

#### macOS:

	cunicu completion zsh > $(brew --prefix)/share/zsh/site-functions/_cunicu

You will need to start a new shell for this setup to take effect.


```
cunicu completion zsh [flags]
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
  -v, --verbose int        verbosity level
```

### SEE ALSO

* [cunicu completion](cunicu_completion.md)	 - Generate the autocompletion script for the specified shell

