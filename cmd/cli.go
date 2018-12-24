package cmd

import (
	"fmt"
	"github.com/gqlc/compiler"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
	"text/scanner"
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

// NewCLI returns a compiler.CommandLine implementation.
func NewCLI() *ccli {
	return &ccli{
		root: rootCmd,
	}
}

func (c *ccli) AllowPlugins(prefix string) { pluginPrefix = prefix }

// outFlag represents a generator output flag
type outFlag struct {
	opts   map[string]interface{}
	outDir string
}

func (f *outFlag) Parse(arg string) error {
	if !strings.Contains(arg, ":") {
		f.outDir = arg
		return nil
	}

	genOpts := strings.Split(arg, ":")
	f.outDir = genOpts[1]

	s := scanner.Scanner{
		Error: func(sc *scanner.Scanner, msg string) {
			// TODO: Handle errors
		},
	}
	s.Init(strings.NewReader(genOpts[0]))

	tok := s.Scan()
	for {
		if tok == scanner.EOF {
			return nil
		}
		if tok == ',' {
			tok = s.Scan()
			continue
		}

		key := s.TokenText()

		t := s.Scan()
		switch t {
		case ',':
			f.opts[key] = true
		case '=':
			var val interface{}
			tt := s.Scan()
			valStr := s.TokenText()

			oldV, ok := f.opts[key]
			switch tt {
			case scanner.Int:
				valInt, err := strconv.ParseInt(valStr, 10, 64)
				if err != nil {
					return err
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
					return err
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
						return err
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
				return fmt.Errorf("gqlc: malformed generator option: %s", key)
			}

			f.opts[key] = val
		case scanner.EOF:
			f.opts[key] = true
			return nil
		default:
			return fmt.Errorf("gqlc: unexpected character in generator option: %s", string(t))
		}

		tok = s.Scan()
	}
}

func (*outFlag) String() string { return "" }

func (f *outFlag) Set(s string) error { return f.Parse(s) }

func (*outFlag) Type() string { return "string" }

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

	oFlag := &outFlag{opts: make(map[string]interface{})}
	c.root.Flags().Var(oFlag, name, help)
	geners[name] = g

	if opt != "" {
		c.root.Flags().Var(oFlag, opt, "Pass additional options to generator.")
		opts[opt] = g
	}
}

func (c *ccli) Run(args []string) error {
	return c.root.Execute()
}
