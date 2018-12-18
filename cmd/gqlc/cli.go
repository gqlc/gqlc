package main

import (
	"github.com/Zaba505/gqlc/compiler"
	"github.com/spf13/cobra"
)

var (
	pluginPrefix string
	geners       map[string]compiler.CodeGenerator
)

func init() {
	geners = make(map[string]compiler.CodeGenerator)
}

// ccli is an implementation of the compiler interface, which
// simply wraps a github.com/spf13/cobra.Command
type ccli struct {
	root *cobra.Command
}

func newCLI() *ccli {
	return &ccli{
		root: rootCmd,
	}
}

func (c *ccli) AllowPlugins(prefix string) { pluginPrefix = prefix }

func (c *ccli) RegisterGenerator(name string, g compiler.CodeGenerator, help string) {
	c.root.Flags().String(name, "", help)
	geners[name] = g
}

func (c *ccli) Run(args []string) error {
	return c.root.Execute()
}
