// Package cmd provides a compiler.CommandLine implementation.
package cmd

import (
	"fmt"
	"github.com/gqlc/gqlc/gen"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"runtime/debug"
	"text/scanner"
)

type option func(*CommandLine)

// WithCommand configures the underlying cobra.Command to be used.
func WithCommand(cmd *cobra.Command) option {
	return func(c *CommandLine) {
		c.Command = cmd
	}
}

// WithFS configures the underlying afero.FS used to read/write files.
func WithFS(fs afero.Fs) option {
	return func(c *CommandLine) {
		c.fs = fs
	}
}

// CommandLine which simply extends a github.com/spf13/cobra.Command
// to include helper methods for registering code generators.
//
type CommandLine struct {
	*cobra.Command

	pluginPrefix *string
	geners       []*genFlag
	fp           *fparser
	fs           afero.Fs
}

// NewCLI returns a CommandLine implementation.
func NewCLI(opts ...option) (c *CommandLine) {
	c = &CommandLine{
		pluginPrefix: new(string),
		fp: &fparser{
			Scanner: new(scanner.Scanner),
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.fs == nil {
		c.fs = afero.NewOsFs()
	}

	if c.Command != nil {
		return
	}
	c.Command = rootCmd
	c.PreRunE = chainPreRunEs(
		parseFlags(c.pluginPrefix, &c.geners, c.fp),
		validateArgs,
		validatePluginTypes(c.fs),
		initGenDirs(c.fs, c.geners),
	)
	c.RunE = func(cmd *cobra.Command, args []string) error {
		if len(cmd.Flags().Args()) == 0 || cmd.Flags().Lookup("help").Changed {
			return cmd.Help()
		}

		importPaths, err := cmd.Flags().GetStringSlice("import_path")
		if err != nil {
			return err
		}

		return root(c.fs, &c.geners, importPaths, cmd.Flags().Args()...)
	}

	return
}

func (c *CommandLine) AllowPlugins(prefix string) { *c.pluginPrefix = prefix }

func (c *CommandLine) RegisterGenerator(g gen.Generator, details ...string) {
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

func wrapPanic(err error, stack []byte) error {
	return fmt.Errorf("gqlc: recovered from unexpected panic: %w\n\n%s", err, stack)
}

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

	c.SetArgs(args[1:])
	return c.Execute()
}
