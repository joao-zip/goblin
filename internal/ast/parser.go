// Package ast provides utilities for parsing and manipulating Go source files.
package ast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// ParseSource parses Go source code from a string.
func ParseSource(src string) (*ast.File, *token.FileSet, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "source.go", src, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing source: %w", err)
	}
	return file, fset, nil
}

// ParseFile parses a single Go source file from disk.
func ParseFile(path string) (*ast.File, *token.FileSet, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing file %s: %w", path, err)
	}
	return file, fset, nil
}

// ParseDir parses all non-test Go files in a directory.
func ParseDir(dir string) ([]*ast.File, *token.FileSet, error) {
	fset := token.NewFileSet()

	filter := func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}

	pkgs, err := parser.ParseDir(fset, dir, filter, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing directory %s: %w", dir, err)
	}

	var files []*ast.File
	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			files = append(files, f)
		}
	}

	return files, fset, nil
}
