package config

import "github.com/spf13/cobra"

type ValidArgsFunction func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)

var BooleanCompletions = cobra.FixedCompletions([]string{"true", "false"}, cobra.ShellCompDirectiveNoFileComp)
