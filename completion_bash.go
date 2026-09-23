package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func completionBashCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "bash",
		Short: "Generate the autocompletion script for bash",
		Long: `Generate the autocompletion script for the bash shell.

This aliases roam completion to git completion, since roam wraps git.
It needs git's bash completion, usually provided by the 'bash-completion'
package. If it is not installed already, you can install it via your OS's
package manager.

To load completions in your current shell session:

	source <(roam completion bash)

To load completions for every new session, execute once:

#### Linux:

	roam completion bash > /etc/bash_completion.d/roam

#### macOS:

	roam completion bash > $(brew --prefix)/etc/bash_completion.d/roam

You will need to start a new shell for this setup to take effect.`,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "complete -F _git roam")
			return err
		},
	}
}
