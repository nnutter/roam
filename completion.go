package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func completionCommand(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "completion",
		Short:         "Generate the autocompletion script for the specified shell",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return nil
			}
			return fmt.Errorf("unknown generator: %s", args[0])
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(completionBashCommand(), completionZshCommand(a))
	return cmd
}
