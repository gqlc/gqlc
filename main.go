// Command gqlc is a multi-language GraphQL implementation generator.
package main

import (
	"fmt"
	"github.com/gqlc/compiler"
	"github.com/gqlc/doc"
	"github.com/gqlc/golang"
	"github.com/gqlc/gqlc/cmd"
	"github.com/gqlc/js"
	"os"
)

var cli compiler.CommandLine

func init() {
	cli = cmd.NewCLI()
	cli.AllowPlugins("gqlc-gen-")

	// Register Documentation generator
	cli.RegisterGenerator(&doc.Generator{},
		"doc_out",
		"doc_opt",
		"Generate Documentation from GraphQL schema.",
	)

	// Register Go generator
	cli.RegisterGenerator(&golang.Generator{},
		"go_out",
		"go_opt",
		"Generate Go source.",
	)

	// Register Javascript generator
	cli.RegisterGenerator(&js.Generator{},
		"js_out",
		"js_opt",
		"Generate Javascript source.",
	)
}

func main() {
	if err := cli.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
