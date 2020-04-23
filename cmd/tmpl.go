package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"strings"
	"text/template"
)

const usageTmpl = `Usage:
  gqlc [flags] files
  gqlc [command]{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{$flags := filter .LocalFlags "_opt" false}}{{$inflags := filter $flags "_out" true}}{{if gt (len $inflags.FlagUsages) 0}}

Generator Flags:
{{$inflags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{$exflags := filter $flags "_out" false}}{{if gt (len $exflags.FlagUsages) 0}}

General Flags:
{{$exflags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasExample}}

Example:
	{{.Example}}{{end}}
`

func filterFlags(set *pflag.FlagSet, key string, ex bool) *pflag.FlagSet {
	fs := new(pflag.FlagSet)
	set.VisitAll(func(flag *pflag.Flag) {
		if strings.Contains(flag.Name, key) == ex {
			fs.AddFlag(flag)
		}
	})
	return fs
}

func init() {
	cobra.AddTemplateFuncs(template.FuncMap{
		"filter": filterFlags,
	})
}
