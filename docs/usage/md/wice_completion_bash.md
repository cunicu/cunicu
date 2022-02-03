## wice completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(wice completion bash)

To load completions for every new session, execute once:

#### Linux:

	wice completion bash > /etc/bash_completion.d/wice

#### macOS:

	wice completion bash > /usr/local/etc/bash_completion.d/wice

You will need to start a new shell for this setup to take effect.


```
wice completion bash
```

### Options

```
  -h, --help              help for bash
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
  -d, --log-level string   log level (one of "debug", "info", "warn", "error", "dpanic", "panic", and "fatal") (default "info")
```

### SEE ALSO

* [wice completion](wice_completion.md)	 - Generate the autocompletion script for the specified shell

