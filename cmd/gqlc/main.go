package main

import (
	"fmt"
	"github.com/Zaba505/gqlc/compiler"
	"github.com/Zaba505/gqlc/compiler/dart"
	"github.com/Zaba505/gqlc/compiler/doc"
	"github.com/Zaba505/gqlc/compiler/golang"
	"github.com/Zaba505/gqlc/compiler/js"
	"os"
)

var cli compiler.CommandLine

func init() {
	cli = newCLI()
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
