// Command gqlc is a multi-language GraphQL implementation generator.
package main

import (
	"fmt"
	"github.com/gqlc/compiler"
	"github.com/gqlc/compiler/dart"
	"github.com/gqlc/compiler/doc"
	"github.com/gqlc/compiler/golang"
	"github.com/gqlc/compiler/js"
	"github.com/gqlc/gqlc/cmd"
	"os"
)

var cli compiler.CommandLine

func init() {
	cli = cmd.NewCLI()
	cli.AllowPlugins("gqlc-gen-")

	// Register Dart generator
	cli.RegisterGenerator(&dart.Generator{},
		"dart_out",
		"Generate Dart source.",
	)

	// Register Documentation generator
	cli.RegisterGenerator(&doc.Generator{},
		"doc_out",
		"doc_opt",
		"Generate Documentation from GraphQL schema.",
	)

	// Register Go generator
	cli.RegisterGenerator(&golang.Generator{},
		"go_out",
		"Generate Go source.",
	)

	// Register Javascript generator
	cli.RegisterGenerator(&js.Generator{},
		"js_out",
		"Generate Javascript source.",
	)
}

func main() {
	if err := cli.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
