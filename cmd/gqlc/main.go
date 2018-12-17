package main

import (
	"fmt"
	"github.com/Zaba505/gqlc/compiler"
	"github.com/Zaba505/gqlc/compiler/dart"
	"github.com/Zaba505/gqlc/compiler/doc"
	"github.com/Zaba505/gqlc/compiler/go"
	"github.com/Zaba505/gqlc/compiler/js"
	"os"
)

var cli compiler.CommandLine

func init() {
	cli = NewCLI()
	cli.AllowPlugins("gqlc-gen-")

	// Register Documentation generator
	cli.RegisterGenerator("doc_out", &doc.Generator{},
		"Generate Documentation from GraphQL schema.")

	// Register Go generator
	cli.RegisterGenerator("go_out", &_go.Generator{},
		"Generate Go source.")

	// Register Javascript generator
	cli.RegisterGenerator("js_out", &js.Generator{},
		"Generate Javascript source.")

	// Register Dart generator
	cli.RegisterGenerator("dart_out", &dart.Generator{},
		"Generate Dart source.")
}

func main() {
	if err := cli.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
