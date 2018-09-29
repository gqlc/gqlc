package translator

// CodeGenerator provides a simple API for creating code generator for
// any language desired
type CodeGenerator interface {
	Generate()
	GenerateAll()
}