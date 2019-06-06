// Package cmd provides a compiler.CommandLine implementation.
package cmd

import (
	"fmt"
	"github.com/gqlc/compiler"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"text/scanner"
)

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
func NewCLI() *cli {
	c := &cli{
		Command:      rootCmd,
		pluginPrefix: new(string),
		fp: &fparser{
			Scanner: new(scanner.Scanner),
		},
	}

	fs := afero.NewOsFs()
	c.PreRunE = chainPreRunEs(
		parseFlags(c.pluginPrefix, &c.geners, c.fp),
		validateArgs,
		validatePluginTypes(fs),
		initGenDirs(fs, c.geners),
	)
	c.RunE = root(fs, &c.geners)
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

// Flags

// genFlag represents a Generator flag: *_out
type genFlag struct {
	compiler.Generator
	outDir *string
	opts   map[string]interface{}

	geners *[]*genFlag
	fp     *fparser
}

func (*genFlag) String() string { return "" }

func (*genFlag) Type() string { return "string" }

func (f *genFlag) Set(arg string) (err error) {
	*f.geners = append(*f.geners, f)
	f.fp.Init(strings.NewReader(arg))

	err = f.fp.parse(parseArg, f.outDir, f.opts)
	if err != nil {
		return err
	}

	if *f.outDir != "." {
		return
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	*f.outDir = wd
	return
}

// genOptFlag represents a Generator option flag: *_opt
type genOptFlag struct {
	opts map[string]interface{}
	fp   *fparser
}

func (*genOptFlag) String() string { return "" }

func (*genOptFlag) Type() string { return "string" }

func (f *genOptFlag) Set(arg string) error {
	f.fp.Init(strings.NewReader(arg))
	return f.fp.parse(parseArg, nil, f.opts)
}

type stateFn func(*fparser, *string, map[string]interface{}) stateFn

type fparser struct {
	*scanner.Scanner
}

func (p *fparser) errorf(format string, args ...interface{}) { panic(fmt.Errorf(format, args...)) }

func (p *fparser) error(err error) { panic(err) }

func (p *fparser) recover(err *error) {
	e := recover()
	if e != nil {
		*err = e.(error)
	}
}

func (p *fparser) parse(root stateFn, dir *string, opts map[string]interface{}) (err error) {
	defer p.recover(&err)

	for state := root; state != nil; {
		state = state(p, dir, opts)
	}
	return
}

func parseArg(p *fparser, dir *string, opts map[string]interface{}) stateFn {
	switch t := p.Scan(); t {
	case os.PathSeparator:
		*dir += string(os.PathSeparator)
		return parseDir(p, dir)
	case '.':
		*dir = "."
		return nil
	}

	key := p.TokenText()

	switch tt := p.Scan(); tt {
	case ':':
		fallthrough
	case ',':
		opts[key] = true
		return parseArg
	case '=':
		return parseValue(key)
	case os.PathSeparator:
		*dir = *dir + key + string(os.PathSeparator)
		return parseDir(p, dir)
	case scanner.EOF:
		if dir != nil {
			*dir = key
			return nil
		}
		if key != "" {
			opts[key] = true
		}
	}

	return nil
}

func parseValue(key string) stateFn {
	return func(p *fparser, dir *string, opts map[string]interface{}) stateFn {
		var val interface{}
		tt := p.Scan()
		valStr := p.TokenText()

		oldV, ok := opts[key]
		switch tt {
		case scanner.Int:
			valInt, err := strconv.ParseInt(valStr, 10, 64)
			if err != nil {
				p.error(err)
			}

			if !ok {
				val = valInt
				break
			}

			oldS, isS := oldV.([]int64)
			if !isS {
				val = append(oldS, oldV.(int64), valInt)
				break
			}
			val = append(oldS, valInt)
		case scanner.Float:
			valFloat, err := strconv.ParseFloat(valStr, 64)
			if err != nil {
				p.error(err)
			}

			if !ok {
				val = valFloat
				break
			}

			oldS, isS := oldV.([]float64)
			if !isS {
				val = append(oldS, oldV.(float64), valFloat)
				break
			}
			val = append(oldS, valFloat)
		case scanner.Ident:
			if valStr == "true" || valStr == "false" {
				valBool, err := strconv.ParseBool(valStr)
				if err != nil {
					p.error(err)
				}

				if !ok {
					val = valBool
					break
				}

				oldS, isS := oldV.([]bool)
				if !isS {
					val = append(oldS, oldV.(bool), valBool)
					break
				}
				val = append(oldS, valBool)
				break
			}

			fallthrough
		case scanner.String:
			if !ok {
				val = valStr
				break
			}

			oldS, isS := oldV.([]string)
			if !isS {
				val = append(oldS, oldV.(string), valStr)
				break
			}
			val = append(oldS, valStr)
		default:
			p.errorf("gqlc: unexpected character in generator option, %s, value: %s", key, string(tt))
		}

		opts[key] = val
		if t := p.Scan(); t == ':' {
			return parseDir(p, dir)
		}

		return parseArg
	}
}

func parseDir(p *fparser, dir *string) stateFn {
	for t := p.Scan(); t != scanner.EOF; {
		*dir += p.TokenText()
		t = p.Scan()
	}
	return nil
}
