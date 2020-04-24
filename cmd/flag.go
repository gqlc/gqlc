package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/scanner"

	"github.com/gqlc/gqlc/gen"
)

// genFlag represents a Generator flag: *_out
type genFlag struct {
	g    gen.Generator
	opts map[string]interface{}

	geners  *[]generator
	outDirs *[]string
	fp      *fparser

	isOpt bool
}

func (genFlag) String() string { return "" }

func (genFlag) Type() string { return "string" }

func (f genFlag) Set(arg string) (err error) {
	if f.isOpt {
		f.fp.Init(strings.NewReader(arg))
		return f.fp.parse(parseArg, nil, f.opts)
	}
	outDir := new(string)

	f.fp.Init(strings.NewReader(arg))

	err = f.fp.parse(parseArg, outDir, f.opts)
	if err != nil {
		return err
	}

	if !filepath.IsAbs(*outDir) {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		*outDir = filepath.Join(wd, *outDir)
	}

	*f.outDirs = append(*f.outDirs, *outDir)
	*f.geners = append(*f.geners, generator{Generator: f.g, opts: f.opts, outDir: *outDir})
	return
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
		t = p.Peek()
		if t == '.' || t == '/' {
			*dir += p.TokenText()
			return parseDir(p, dir)
		}

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
