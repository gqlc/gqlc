package cmd

import "github.com/spf13/cobra"

var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "Format a GraphQL schema file",
}

func init() {
	rootCmd.AddCommand(fmtCmd)
}
