package main

import (
	"fmt"
	"gqlc/compiler"
	"gqlc/compiler/gens"
	"os"
)

var cli compiler.CommandLine

func init() {
	cli = compiler.NewCLI()
	cli.AllowPlugins("gqlc-")

	// Register go generator
	cli.RegisterGenerator("--go_out", gens.GoGenerator{},
		"Generate Go source file.")

	// Register Javascript generator
	cli.RegisterGenerator("--js_out", gens.JsGenerator{},
		"Generate Javascript source.")

	// Register Dart generator
	cli.RegisterGenerator("--dart_out", gens.DartGenerator{},
		"Generate Dart source file.")
}

func main() {
	if err := cli.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
