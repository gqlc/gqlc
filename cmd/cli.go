// Package cmd provides a compiler.CommandLine implementation.
package cmd

import (
	"fmt"
	"github.com/gqlc/compiler"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"os"
	"runtime"
	"strconv"
	"strings"
	"text/scanner"
)

// ccli is an implementation of the compiler.CommandLine interface, which
// simply extends a github.com/spf13/cobra.Command
type cli struct {
	*cobra.Command

	pluginPrefix *string
	geners       map[string]compiler.Generator
	opts         map[string]compiler.Generator
	genOpts      map[compiler.Generator]*oFlag
}

// NewCLI returns a compiler.CommandLine implementation.
func NewCLI() *cli {
	c := &cli{
		Command:      rootCmd,
		geners:       make(map[string]compiler.Generator),
		opts:         make(map[string]compiler.Generator),
		genOpts:      make(map[compiler.Generator]*oFlag),
		pluginPrefix: new(string),
	}

	fs := afero.NewOsFs()
	c.PreRunE = chainPreRunEs(
		parseFlags(c.pluginPrefix, c.geners, c.opts),
		validateArgs,
		accumulateGens(c.pluginPrefix, c.geners, c.opts, c.genOpts),
		validatePluginTypes,
		mkGenDirs(fs, c.genOpts),
	)
	c.RunE = runRoot(fs, c.genOpts)
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

	f := &oFlag{opts: make(map[string]interface{}), outDir: new(string)}
	outFlag := *f
	outFlag.isOut = true
	c.Flags().Var(outFlag, name, help)
	c.geners[name] = g

	if opt != "" {
		optFlag := *f
		c.Flags().Var(optFlag, opt, "Pass additional options to generator.")
		c.opts[opt] = g
	}
}

func (c *cli) Run(args []string) error {
	c.SetArgs(args[1:])
	return c.Execute()
}

// oFlag represents a generator output/option flag
type oFlag struct {
	opts   map[string]interface{}
	outDir *string
	isOut  bool
}

func (oFlag) String() string { return "" }

func (oFlag) Type() string { return "string" }

func (f oFlag) Set(arg string) error {
	p := fparser{
		Scanner: &scanner.Scanner{
			Error: func(sc *scanner.Scanner, msg string) {
				// TODO: Handle errors
			},
		},
	}
	p.Init(strings.NewReader(arg))

	return p.parse(parseArg, f)
}

type stateFn func(fparser, oFlag) stateFn

type fparser struct {
	*scanner.Scanner
}

func (p fparser) errorf(format string, args ...interface{}) { panic(fmt.Errorf(format, args...)) }

func (p fparser) error(err error) { panic(err) }

func (p fparser) recover(err *error) {
	e := recover()
	if e != nil {
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
		*err = e.(error)
	}
}

func (p fparser) parse(root stateFn, f oFlag) (err error) {
	defer p.recover(&err)

	for state := root; state != nil; {
		state = state(p, f)
	}
	return
}

func parseArg(p fparser, f oFlag) stateFn {
	switch t := p.Scan(); t {
	case os.PathSeparator:
		*f.outDir += string(os.PathSeparator)
		return parseDir
	case '.':
		wd, err := os.Getwd()
		if err != nil {
			p.error(err)
		}
		*f.outDir = wd
		return nil
	}

	key := p.TokenText()

	switch tt := p.Scan(); tt {
	case ':':
		fallthrough
	case ',':
		f.opts[key] = true
		return parseArg
	case '=':
		return parseValue(key)
	case os.PathSeparator:
		*f.outDir += key + string(os.PathSeparator)
		return parseDir
	case scanner.EOF:
		if f.isOut {
			*f.outDir = key
			return nil
		}
		f.opts[key] = true
	}

	return nil
}

func parseValue(key string) stateFn {
	return func(p fparser, f oFlag) stateFn {
		var val interface{}
		tt := p.Scan()
		valStr := p.TokenText()

		oldV, ok := f.opts[key]
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
				val = valBool
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

		f.opts[key] = val
		if t := p.Scan(); t == ':' {
			return parseDir
		}

		return parseArg
	}
}

func parseDir(p fparser, f oFlag) stateFn {
	for t := p.Scan(); t != scanner.EOF; {
		*f.outDir += p.TokenText()
		t = p.Scan()
	}
	return nil
}
