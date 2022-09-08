## cunicu completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(cunicu completion bash)

To load completions for every new session, execute once:

#### Linux:

	cunicu completion bash > /etc/bash_completion.d/cunicu

#### macOS:

	cunicu completion bash > $(brew --prefix)/etc/bash_completion.d/cunicu

You will need to start a new shell for this setup to take effect.


```
cunicu completion bash
```

### Options

```
  -h, --help              help for bash
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

