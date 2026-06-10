package mutator

import (
	"go/ast"
	"go/token"
)

var comparisonReplacements = map[token.Token][]token.Token{
	token.EQL: {token.NEQ}, // == -> !=
	token.NEQ: {token.EQL}, // != -> ==
	token.LSS: {token.GEQ}, // < -> >=
	token.GTR: {token.LEQ}, // > -> <=
	token.LEQ: {token.GTR}, // <= -> >
	token.GEQ: {token.LSS}, // >= -> <
}

// ComparisonMutator swaps comparison operators.
type ComparisonMutator struct{}

func (m *ComparisonMutator) Name() string { return "comparison" }

func (m *ComparisonMutator) CanMutate(node ast.Node) bool {
	bin, ok := node.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	_, exists := comparisonReplacements[bin.Op]
	return exists
}

func (m *ComparisonMutator) Mutate(node ast.Node) []MutatedNode {
	bin, ok := node.(*ast.BinaryExpr)
	if !ok {
		return nil
	}

	replacements, exists := comparisonReplacements[bin.Op]
	if !exists {
		return nil
	}

	original := bin.Op
	var mutations []MutatedNode

	for _, rep := range replacements {
		rep := rep
		mutations = append(mutations, MutatedNode{
			Original:    original.String(),
			Replacement: rep.String(),
			Apply:       func() { bin.Op = rep },
			Rollback:    func() { bin.Op = original },
		})
	}

	return mutations
}