// Package cmd provides a compiler.CommandLine implementation.
package cmd

import (
	"fmt"
	"github.com/gqlc/compiler"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"runtime/debug"
	"text/scanner"
)

type option func(*cli)

// WithCommand configures the underlying cobra.Command to be used.
func WithCommand(cmd *cobra.Command) option {
	return func(c *cli) {
		c.Command = cmd
	}
}

// WithPreRunE configures any pre-ran functionality.
func WithPreRunE(preRunE func(*cli) func(*cobra.Command, []string) error) option {
	return func(c *cli) {
		c.PreRunE = preRunE(c)
	}
}

// WithRunE configures the actual CLI functionality.
func WithRunE(runE func(*cli) func(*cobra.Command, []string) error) option {
	return func(c *cli) {
		c.RunE = runE(c)
	}
}

// ProdOptions configures a CLI for production usage.
func ProdOptions() option {
	return func(c *cli) {
		fs := afero.NewOsFs()
		c.PreRunE = chainPreRunEs(
			parseFlags(c.pluginPrefix, &c.geners, c.fp),
			validateArgs,
			validatePluginTypes(fs),
			initGenDirs(fs, c.geners),
		)
		c.RunE = root(fs, &c.geners)
	}
}

// ccli is an implementation of the compiler.CommandLine interface, which
// simply extends a github.com/spf13/cobra.Command
//
type cli struct {
	*cobra.Command

	pluginPrefix *string
	geners       []*genFlag
	fp           *fparser
}

// NewCLI returns a compiler.CommandLine implementation.
func NewCLI(opts ...option) *cli {
	c := &cli{
		Command:      rootCmd,
		pluginPrefix: new(string),
		fp: &fparser{
			Scanner: new(scanner.Scanner),
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *cli) AllowPlugins(prefix string) { *c.pluginPrefix = prefix }

func (c *cli) RegisterGenerator(g compiler.Generator, details ...string) {
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

	opts := make(map[string]interface{})

	c.Flags().Var(&genFlag{
		Generator: g,
		outDir:    new(string),
		opts:      opts,
		geners:    &c.geners,
		fp:        c.fp,
	}, name, help)

	if opt != "" {
		c.Flags().Var(&genOptFlag{opts: opts, fp: c.fp}, opt, "Pass additional options to generator.")
	}
}

type panicErr struct {
	Err        error
	StackTrace []byte
}

func (e *panicErr) Error() string {
	return fmt.Sprintf("lambda: recovered from unexpected panic: %s\n%s", e.Err, e.StackTrace)
}

func (c *cli) Run(args []string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			rerr, ok := r.(error)
			if ok {
				err = &panicErr{Err: rerr, StackTrace: debug.Stack()}
				return
			}

			err = &panicErr{Err: fmt.Errorf("%#v", r), StackTrace: debug.Stack()}
		}
	}()

	c.SetArgs(args[1:])
	return c.Execute()
}
