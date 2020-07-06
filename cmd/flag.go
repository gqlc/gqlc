package cmd

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/scanner"

	"github.com/gqlc/gqlc/gen"
)

// headerFlag represents a flag for setting HTTP headers
// Any repeats will not override. They will append.
//
// format: a=1,b=2
//
type headerFlag struct {
	value   *http.Header
	changed bool
}

func (*headerFlag) String() string { return "" }

func (*headerFlag) Type() string { return "map[string][]string" }

func (f *headerFlag) Set(val string) error {
	var ss []string
	n := strings.Count(val, "=")
	switch n {
	case 0:
		return fmt.Errorf("%s must be formatted as key=value", val)
	case 1:
		ss = append(ss, strings.Trim(val, `"`))
	default:
		r := csv.NewReader(strings.NewReader(val))
		var err error
		ss, err = r.Read()
		if err != nil {
			return err
		}
	}

	out := make(http.Header, len(ss))
	for _, pair := range ss {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			return fmt.Errorf("%s must be formatted as key=value", pair)
		}
		out.Add(kv[0], strings.Trim(kv[1], "\""))
	}
	if !f.changed {
		*f.value = out
	} else {
		for k, v := range out {
			for _, s := range v {
				f.value.Add(k, strings.Trim(s, "\""))
			}
		}
	}
	f.changed = true
	return nil
}

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
		return parseLongForm(key)
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

func parseLongForm(key string) stateFn {
	return func(p *fparser, dir *string, opts map[string]interface{}) stateFn {
		opts[key] = parseValue(p, key, opts)
		if t := p.Scan(); t == ':' {
			return parseDir(p, dir)
		}

		return parseArg
	}
}

func parseValue(p *fparser, key string, opts map[string]interface{}) (val interface{}) {
	tt := p.Scan()
	switch tt {
	case scanner.Int:
		return parseInt(p, opts, key)
	case scanner.Float:
		return parseFloat(p, opts, key)
	case scanner.Ident:
		valStr := p.TokenText()
		if valStr == "true" || valStr == "false" {
			return parseBool(p, opts, key)
		}

		fallthrough
	case scanner.String:
		valStr := p.TokenText()
		oldV, ok := opts[key]
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
	return val
}

func parseInt(p *fparser, opts map[string]interface{}, key string) (val interface{}) {
	valStr := p.TokenText()
	oldV, ok := opts[key]

	valInt, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		p.error(err)
	}

	if !ok {
		val = valInt
		return
	}

	oldS, isS := oldV.([]int64)
	if !isS {
		val = append(oldS, oldV.(int64), valInt)
		return
	}
	val = append(oldS, valInt)
	return
}

func parseFloat(p *fparser, opts map[string]interface{}, key string) (val interface{}) {
	valStr := p.TokenText()
	oldV, ok := opts[key]

	valFloat, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		p.error(err)
	}

	if !ok {
		val = valFloat
		return
	}

	oldS, isS := oldV.([]float64)
	if !isS {
		val = append(oldS, oldV.(float64), valFloat)
		return
	}
	val = append(oldS, valFloat)
	return
}

func parseBool(p *fparser, opts map[string]interface{}, key string) (val interface{}) {
	valStr := p.TokenText()
	oldV, ok := opts[key]

	valBool, err := strconv.ParseBool(valStr)
	if err != nil {
		p.error(err)
	}

	if !ok {
		val = valBool
		return
	}

	oldS, isS := oldV.([]bool)
	if !isS {
		val = append(oldS, oldV.(bool), valBool)
		return
	}
	val = append(oldS, valBool)
	return
}

func parseDir(p *fparser, dir *string) stateFn {
	for t := p.Scan(); t != scanner.EOF; {
		*dir += p.TokenText()
		t = p.Scan()
	}
	return nil
}
