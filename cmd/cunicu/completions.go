package main

import "github.com/spf13/cobra"

var BooleanCompletions = cobra.FixedCompletions([]string{"true", "false"}, cobra.ShellCompDirectiveNoFileComp)
