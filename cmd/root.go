package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gqlc/compiler"
	"path/filepath"
	"strings"
)

var rootCmd = &cobra.Command{
	Use:   "gqlc",
	Short: "A GraphQL IDL compiler",
	Long:  ``,
	RunE:  runRoot,
}

var tmplFs = map[string]interface{}{
	"in": func(set *pflag.FlagSet, key string) *pflag.FlagSet {
		fs := new(pflag.FlagSet)
		set.VisitAll(func(flag *pflag.Flag) {
			if strings.Contains(flag.Name, key) {
				fs.AddFlag(flag)
			}
		})
		return fs
	},
	"ex": func(set *pflag.FlagSet, key string) *pflag.FlagSet {
		fs := new(pflag.FlagSet)
		set.VisitAll(func(flag *pflag.Flag) {
			if !strings.Contains(flag.Name, key) {
				fs.AddFlag(flag)
			}
		})
		return fs
	},
}

func init() {
	cobra.AddTemplateFuncs(tmplFs)

	rootCmd.Flags().String("I", ".", "Path")
	rootCmd.Flags().BoolP("verbose", "v", false, "Output for info")
	rootCmd.SetUsageTemplate(`Usage:
	gqlc [flags] files

Generator Flags:{{$flags := in .LocalFlags "_out"}}
{{$flags.FlagUsages | trimTrailingWhitespaces}}

General Flags:{{$flags = ex .LocalFlags "_out"}}
{{$flags.FlagUsages | trimTrailingWhitespaces}}
`)

	// TODO: Add sub commands to template
}

func runRoot(cmd *cobra.Command, args []string) (err error) {
	if len(args) == 0 {
		return errors.New("gqlc: no files provided")
	}

	// Validate file names
	for _, fileName := range args {
		ext := filepath.Ext(fileName)
		if ext != "gql" && ext != "graphql" {
			return fmt.Errorf("gqlc: invalid file extension: %s", fileName)
		}
	}

	// Parse files
	// TODO: Add parser code here

	// Accumulate selected code generators
	var gs []compiler.CodeGenerator
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		gen, exists := gens[f.Name]
		if exists {
			gs = append(gs, gen)
		}
	})

	// Run code generators
	for _, g := range gs {
		err = g.Generate(context.TODO()) // TODO: Replace context
		if err != nil {
			return
		}
	}

	return
}
