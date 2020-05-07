// Command gqlc is a multi-language GraphQL implementation generator.
package main

import (
	"os"

	"github.com/gqlc/gqlc/cmd"
	"github.com/gqlc/gqlc/doc"
	"github.com/gqlc/gqlc/golang"
	"github.com/gqlc/gqlc/js"
	"go.uber.org/zap"
)

func main() {
	cli := cmd.NewCLI()
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

	if err := cli.Run(os.Args); err != nil {
		l, _ := zap.NewDevelopment()
		l.Sugar().Fatal(err)
	}
}
