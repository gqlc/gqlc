package cmd

import (
	"github.com/spf13/cobra"
)

var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "Format a GraphQL schema file(s).",
	Args:  cobra.MinimumNArgs(1),
	RunE:  format,
}

func init() {
	rootCmd.AddCommand(fmtCmd)

	fmtCmd.SetUsageTemplate(`Usage:
	gqlc fmt [flags] files

Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
`)
}

// TODO
func format(cmd *cobra.Command, args []string) error {
	// Parse files
	// Apply format rules
	// Rewrite files
	return nil
}
