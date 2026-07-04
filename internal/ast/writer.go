package ast

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"

	"github.com/joao-zip/goblin/internal/mutator"
)

// WriteFile writes an AST back to a Go source file, formatted with gofmt.
func WriteFile(fset *token.FileSet, file *ast.File, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file %s: %w", path, err)
	}
	defer f.Close()

	if err := format.Node(f, fset, file); err != nil {
		return fmt.Errorf("formatting AST to %s: %w", path, err)
	}
	return nil
}

// ApplyAndWrite applies a mutation, writes the modified AST, then rolls back.
func ApplyAndWrite(fset *token.FileSet, file *ast.File, mut mutator.MutatedNode, path string) error {
	mut.Apply()
	err := WriteFile(fset, file, path)
	mut.Rollback()

	if err != nil {
		return fmt.Errorf("writing mutated file: %w", err)
	}
	return nil
}
