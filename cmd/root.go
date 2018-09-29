package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gqlt/translator"
	"os"
)

var cli = translator.NewCLI()

var rootCmd = &cobra.Command{
	Use: "gqlt",
	Short: "gqlt is a GraphQL translator",
	Long: `A simple translator from the GraphQL language spec to
			concrete types in various languages.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.Flags().String("go_out", ".", "Specify output directory for generated Go code")
	rootCmd.Flags().String("js_out", ".", "Specify output directory for generated Javascript code")
	rootCmd.Flags().String("dart_out", ".", "Specify output directory for generated Dart code")

	viper.BindPFlag("go_out", rootCmd.Flags().Lookup("go_out"))
	viper.BindPFlag("js_out", rootCmd.Flags().Lookup("js_out"))
	viper.BindPFlag("dart_out", rootCmd.Flags().Lookup("dart_out"))

	cli.RegisterGenerator() // TODO: Register Go generator
	cli.RegisterGenerator() // TODO: Register JS generator
	cli.RegisterGenerator() // TODO: Register Dart generator
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}