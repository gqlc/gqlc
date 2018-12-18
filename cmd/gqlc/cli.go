package main

import (
	"github.com/Zaba505/gqlc/compiler"
	"github.com/spf13/cobra"
)

var (
	pluginPrefix string
	geners       map[string]compiler.CodeGenerator
	opts         map[string]compiler.CodeGenerator
)

func init() {
	geners = make(map[string]compiler.CodeGenerator)
	opts = make(map[string]compiler.CodeGenerator)
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

func (c *ccli) RegisterGenerator(g compiler.CodeGenerator, details ...string) {
	l := len(details)
	var name, opt, help string
	switch {
	case l == 2:
		name, help = details[0], details[1]
	case l > 3:
		fallthrough
	case l == 3:
		name, opt, help = details[0], details[1], details[2]
	default:
		panic("invalid generator flag details")
	}

	c.root.Flags().String(name, "", help)
	geners[name] = g

	if opt != "" {
		c.root.Flags().StringSlice(opt, nil, "Pass additional options to generator.")
		opts[opt] = g
	}
}

func (c *ccli) Run(args []string) error {
	return c.root.Execute()
}
