package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gqlc/compiler"
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
		ext := filepath.Ext(fileName)
		if ext != ".gql" && ext != ".graphql" {
			return fmt.Errorf("gqlc: invalid file extension: %s", fileName)
		}
	}

	return nil
}

func init() {
	rootCmd.Flags().StringSliceP("types", "t", nil, "Provide .gql files containing types you wish to register with the compiler.")
}

// validatePluginTypes parses and validates any types given by the --types flag.
func validatePluginTypes(fs afero.Fs) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		pluginTypes, _ := cmd.Flags().GetStringSlice("types")
		if len(pluginTypes) == 0 {
			return nil
		}

		importPaths, err := cmd.Flags().GetStringSlice("import_path")
		if err != nil {
			return err
		}

		docMap := make(map[string]*ast.Document, len(pluginTypes))
		err = parseInputFiles(fs, token.NewDocSet(), docMap, importPaths, pluginTypes...)
		if err != nil {
			return err
		}

		docs := make([]*ast.Document, 0, len(docMap))
		for _, doc := range docMap {
			docs = append(docs, doc)
		}

		docsIR, err := compiler.ReduceImports(docs)
		if err != nil {
			return err
		}

		errs := compiler.CheckTypes(docsIR, compiler.TypeCheckerFn(compiler.Validate))
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
