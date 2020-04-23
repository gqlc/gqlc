package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version is version of the current gqlc binary.
// dev refers to a local build. It will be overwritten
// during CI/CD.
//
var version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("gqlc %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
