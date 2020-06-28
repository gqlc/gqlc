package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gqlc/compiler"
	"github.com/gqlc/compiler/spec"
	"github.com/gqlc/graphql/ast"
	"github.com/gqlc/graphql/token"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func chainPreRunEs(preRunEs ...func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		for i := 0; i < len(preRunEs) && err == nil; i++ {
			err = preRunEs[i](cmd, args)
		}
		return
	}
}

// validateFilenames validates that only GraphQL files are provided.
func validateFilenames(cmd *cobra.Command, args []string) error {
	for _, fileName := range args {
		if strings.HasPrefix(fileName, "http") || strings.HasPrefix(fileName, "ws") {
			continue
		}

		ext := filepath.Ext(fileName)
		if ext != ".gql" && ext != ".graphql" {
			return fmt.Errorf("gqlc: invalid file extension: %s", fileName)
		}
	}

	return nil
}

// validatePluginTypes parses and validates any types given by the --types flag.
func (c *gqlcCmd) validatePluginTypes(fs afero.Fs) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		pluginTypes, _ := cmd.Flags().GetStringSlice("types")
		if len(pluginTypes) == 0 {
			return nil
		}

		docMap := make(map[string]*ast.Document, len(pluginTypes))
		err := c.parseInputFiles(fs, token.NewDocSet(), docMap, pluginTypes...)
		if err != nil {
			return err
		}

		docs := make([]*ast.Document, 0, len(docMap))
		for _, doc := range docMap {
			docs = append(docs, doc)
		}

		docsIR := compiler.ToIR(docs)

		docsIR, err = compiler.ReduceImports(docsIR)
		if err != nil {
			return err
		}

		errs := compiler.CheckTypes(docsIR, spec.Validator, compiler.ImportValidator)
		if len(errs) > 0 {
			// TODO: Compound errs
			return nil
		}

		return nil
	}
}

// initGenDirs initializes each directory each generator will be outputting to.
func initGenDirs(fs afero.Fs, dirs *[]string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		for _, dir := range *dirs {
			zap.S().Info("creating directory:", dir)
			err = fs.MkdirAll(dir, os.ModeDir)
			if err != nil {
				break
			}
		}
		return
	}
}
