package compiler

// CommandLine provides a clean and concise way to implement
// CLIs for compiling GraphQL Schema Language
type CommandLine interface {
	// RegisterGenerator register a language generator with CLI
	RegisterGenerator(flagName string, gen CodeGenerator, helpText string)

	// AllowPlugins enables "plugins". If a command-line flag ends with "_out"
	// but does not match any register code generator, the compiler will
	// attempt to find the "plugin" to implement the generator. Plugins are
	// just executables. They should reside in your PATH.
	//
	// The compiler determines the executable name to search for by concatenating
	// exe_name_prefix with the unrecognized flag name, removing "_out".  So, for
	// example, if exe_name_prefix is "protoc-" and you pass the flag --foo_out,
	// the compiler will try to run the program "protoc-foo".
	//
	// The plugin program should implement the following usage:
	//   plugin [--out=OUTDIR] [--parameter=PARAMETER] PROTO_FILES < DESCRIPTORS
	// --out indicates the output directory (as passed to the --foo_out
	// parameter); if omitted, the current directory should be used.  --parameter
	// gives the generator parameter, if any was provided (see below).  The
	// PROTO_FILES list the .proto files which were given on the compiler
	// command-line; these are the files for which the plugin is expected to
	// generate output code.  Finally, DESCRIPTORS is an encoded FileDescriptorSet
	// (as defined in descriptor.proto).  This is piped to the plugin's stdin.
	// The set will include descriptors for all the files listed in PROTO_FILES as
	// well as all files that they import.  The plugin MUST NOT attempt to read
	// the PROTO_FILES directly -- it must use the FileDescriptorSet.
	//
	// The plugin should generate whatever files are necessary, as code generators
	// normally do.  It should write the names of all files it generates to
	// stdout.  The names should be relative to the output directory, NOT absolute
	// names or relative to the current directory.  If any errors occur, error
	// messages should be written to stderr.  If an error is fatal, the plugin
	// should exit with a non-zero exit code.
	AllowPlugins(exeNamePrefix string)

	// Run the compiler with the given command-line parameters.
	Run(args []string) error
}

type cli struct{}

func (c cli) RegisterGenerator(flagName string, gen CodeGenerator, helpText string) {}

func (c cli) AllowPlugins(exeNamePrefix string) {}

func (c cli) Run(args []string) error {
	return nil
}

// NewCLI returns a simple implementation for the CommandLine interface
func NewCLI() CommandLine {
	return cli{}
}
