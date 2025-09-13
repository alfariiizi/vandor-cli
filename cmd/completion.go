package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for Vandor CLI.

The completion script can be sourced directly or saved to a file and sourced later.

Examples:
  # Generate and source bash completion (Linux/macOS)
  source <(vandor completion bash)

  # Generate and source zsh completion
  source <(vandor completion zsh)

  # Generate and source fish completion
  vandor completion fish | source

  # Generate powershell completion
  vandor completion powershell | Out-String | Invoke-Expression

  # Save bash completion to file
  vandor completion bash > ~/.local/share/bash-completion/completions/vandor

  # Save zsh completion to file (oh-my-zsh)
  vandor completion zsh > ~/.oh-my-zsh/completions/_vandor

  # Save zsh completion to file (standard location)
  vandor completion zsh > /usr/local/share/zsh/site-functions/_vandor

  # Save fish completion to file
  vandor completion fish > ~/.config/fish/completions/vandor.fish

Installation Instructions:

  Bash:
    # Add to ~/.bashrc or ~/.bash_profile:
    source <(vandor completion bash)

    # Or save to completion directory:
    vandor completion bash > ~/.local/share/bash-completion/completions/vandor

  Zsh:
    # Make sure completion is enabled in ~/.zshrc:
    autoload -U compinit; compinit

    # Then add one of these to ~/.zshrc:
    source <(vandor completion zsh)

    # Or for oh-my-zsh users:
    vandor completion zsh > ~/.oh-my-zsh/completions/_vandor

    # Or system-wide:
    sudo vandor completion zsh > /usr/local/share/zsh/site-functions/_vandor

  Fish:
    vandor completion fish > ~/.config/fish/completions/vandor.fish

  PowerShell:
    # Add to your PowerShell profile:
    vandor completion powershell | Out-String | Invoke-Expression

Note: When you upgrade Vandor CLI using 'vandor upgrade', completions will be
automatically regenerated if they were previously installed in standard locations.`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			err := cmd.Root().GenBashCompletion(os.Stdout)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating bash completion: %v\n", err)
				os.Exit(1)
			}
		case "zsh":
			err := cmd.Root().GenZshCompletion(os.Stdout)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating zsh completion: %v\n", err)
				os.Exit(1)
			}
		case "fish":
			err := cmd.Root().GenFishCompletion(os.Stdout, true)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating fish completion: %v\n", err)
				os.Exit(1)
			}
		case "powershell":
			err := cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating powershell completion: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}