package ast

import "github.com/Zaba505/gqlc/graphql/token"

// SortImports sorts runs of consecutive import lines in import blocks in d.
// It also removes duplicate imports when it is possible to do so without data loss.
func SortImports(dset *token.DocSet, d *token.Doc) {}
