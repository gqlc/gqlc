// Package cmd implements the command line interface for gqlc.
package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/gqlc/gqlc/gen"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type option func(*CommandLine)

// WithFS configures the underlying afero.FS used to read/write files.
func WithFS(fs afero.Fs) option {
	return func(c *CommandLine) {
		c.fs = fs
	}
}

type genConfig struct {
	g    gen.Generator
	name string
	opt  string
	help string
}

// CommandLine provides a convient API for adding generators to gqlc.
type CommandLine struct {
	prefix string
	fs     afero.Fs

	cmds []cmder
	gens []genConfig
}

type cmder interface {
	getCommand() *cobra.Command
}

type baseCmd struct {
	*cobra.Command
}

func (cmd *baseCmd) getCommand() *cobra.Command { return cmd.Command }

func (c *CommandLine) addCommand(cmds ...cmder) *CommandLine {
	c.cmds = append(c.cmds, cmds...)
	return c
}

func (c *CommandLine) build() *cobra.Command {
	cmd := c.newGqlcCmd(c.gens, c.fs, c.prefix)
	for _, cmdr := range c.cmds {
		cmd.AddCommand(cmdr.getCommand())
	}

	return cmd.Command
}

// NewCLI returns a CommandLine implementation.
func NewCLI(opts ...option) (c *CommandLine) {
	c = new(CommandLine)

	for _, opt := range opts {
		opt(c)
	}

	if c.fs == nil {
		c.fs = afero.NewOsFs()
	}

	return
}

// AllowPlugins sets the plugin prefix to be used
// when looking up plugin executables.
//
func (c *CommandLine) AllowPlugins(prefix string) { c.prefix = prefix }

// RegisterGenerator registers a generator with the compiler.
func (c *CommandLine) RegisterGenerator(g gen.Generator, name, opt, help string) {
	c.gens = append(c.gens, genConfig{
		g:    g,
		name: name,
		opt:  opt,
		help: help,
	})
}

func wrapPanic(err error, stack []byte) error {
	return fmt.Errorf("gqlc: recovered from unexpected panic: %w\n\n%s", err, stack)
}

// Run executes the compiler
func (c *CommandLine) Run(args []string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()

			rerr, ok := r.(error)
			if ok {
				err = wrapPanic(rerr, stack)
				return
			}

			err = wrapPanic(fmt.Errorf("%#v", r), stack)
		}
	}()

	cmd := c.addCommand(c.newVersionCmd()).build()

	cmd.SetArgs(args[1:])
	return cmd.Execute()
}
