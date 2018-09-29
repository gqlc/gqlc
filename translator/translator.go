package translator

// CommandLine provides a clean and concise way to implement
// CLIs for translating GraphQL language into any other programming language
type CommandLine interface {
	RegisterGenerator()
	AllowPlugins()
	Run()
}

// NewCLI returns a simple implementation for the CommandLine interface
func NewCLI() CommandLine {
	return nil
}