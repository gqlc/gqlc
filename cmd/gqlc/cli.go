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

// CLI is an implementation of the compiler interface, which
// simply wraps a github.com/spf13/cobra.Command
type CLI struct {
	root *cobra.Command
}

func NewCLI() *CLI {
	return &CLI{
		root: rootCmd,
	}
}

func (c *CLI) AllowPlugins(prefix string) { pluginPrefix = prefix }

func (c *CLI) RegisterGenerator(name string, g compiler.CodeGenerator, help string) {
	c.root.Flags().String(name, "", help)
	geners[name] = g
}

func (c *CLI) Run(args []string) error {
	return c.root.Execute()
}
