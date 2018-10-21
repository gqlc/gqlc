package cmd

import (
	"github.com/spf13/cobra"
	"gqlc/compiler"
)

var gens map[string]compiler.CodeGenerator

func init() {
	gens = make(map[string]compiler.CodeGenerator)
}

// CLI is an implementation of the compiler interface
type CLI struct {
	root *cobra.Command
}

func NewCLI() *CLI {
	return &CLI{
		root: rootCmd,
	}
}

func (c *CLI) AllowPlugins(prefix string) {}

func (c *CLI) RegisterGenerator(name string, g compiler.CodeGenerator, help string) {
	c.root.Flags().String(name, "", help)
	gens[name] = g
}

func (c *CLI) Run(args []string) error {
	return c.root.Execute()
}
