package main

import (
	"fmt"
	"github.com/Zaba505/gqlc/cmd"
	"github.com/Zaba505/gqlc/compiler"
	"github.com/Zaba505/gqlc/compiler/gens"
	"os"
)

var cli compiler.CommandLine

func init() {
	cli = cmd.NewCLI()
	cli.AllowPlugins("gqlc-gen-")

	// Register Documentation generator
	cli.RegisterGenerator("doc_out", gens.DocGenerator{},
		"Generate Documentation from GraphQL schema.")

	// Register Go generator
	cli.RegisterGenerator("go_out", gens.GoGenerator{},
		"Generate Go source.")

	// Register Javascript generator
	cli.RegisterGenerator("js_out", gens.JsGenerator{},
		"Generate Javascript source.")

	// Register Dart generator
	cli.RegisterGenerator("dart_out", gens.DartGenerator{},
		"Generate Dart source.")
}

func main() {
	if err := cli.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
