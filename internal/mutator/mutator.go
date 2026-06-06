// Package mutator defines the Mutator interface and concrete mutation strategies.
package mutator

import "go/ast"

// MutatedNode represents a single possible mutation on an AST node.
type MutatedNode struct {
	Original    string
	Replacement string
	Apply       func() // Applies the mutation to the AST
	Rollback    func() // Reverts the mutation
}

// Mutator generates mutations for AST nodes.
type Mutator interface {
	Name() string
	CanMutate(node ast.Node) bool
	Mutate(node ast.Node) []MutatedNode
}
