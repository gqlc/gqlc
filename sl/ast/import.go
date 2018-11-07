package ast

import "gqlc/sl/token"

// SortImports sorts runs of consecutive import lines in import blocks in f.
// It also removes duplicate imports when it is possible to do so without data loss.
func SortImports(fset *token.FileSet, f *File) {}
