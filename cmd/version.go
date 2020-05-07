package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// version is version of the current gqlc binary.
// dev refers to a local build. It will be overwritten
// during CI/CD.
//
var version = "D.E.V"

func (c *CommandLine) newVersionCmd() *baseCmd {
	return &baseCmd{
		Command: &cobra.Command{
			Use:   "version",
			Short: "Print the version",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Printf("gqlc version gqlc%s %s/%s\n", version, runtime.GOOS, runtime.GOARCH)
			},
		},
	}
}
