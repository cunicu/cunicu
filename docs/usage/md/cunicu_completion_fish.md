---
title: cunicu completion fish
sidebar_label: completion fish
sidebar_class_name: command-name
slug: /usage/man/completion/fish
hide_title: true
keywords:
    - manpage
---

## cunicu completion fish

Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	cunicu completion fish | source

To load completions for every new session, execute once:

	cunicu completion fish > ~/.config/fish/completions/cunicu.fish

You will need to start a new shell for this setup to take effect.


```
cunicu completion fish [flags]
```

### Options

```
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
  -q, --color string       Enable colorization of output (one of: auto, always, never) (default "auto")
  -l, --log-file string    path of a file to write logs to
  -d, --log-level string   log level (one of: debug, info, warn, error, dpanic, panic, and fatal) (default "info")
  -v, --verbose int        verbosity level
```

### SEE ALSO

* [cunicu completion](cunicu_completion.md)	 - Generate the autocompletion script for the specified shell

