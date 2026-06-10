package mutator

import (
	"go/ast"
	"go/token"
)


var logicalReplacements = map[token.Token][]token.Token{
	token.LAND: {token.LOR},  // && -> ||
	token.LOR:  {token.LAND}, // || -> &&
}

// LogicalMutator swaps logical operators.
type LogicalMutator struct{}

func (m *LogicalMutator) Name() string { return "logical" }

func (m *LogicalMutator) CanMutate(node ast.Node) bool {
	bin, ok := node.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	_, exists := logicalReplacements[bin.Op]
	return exists
}

func (m *LogicalMutator) Mutate(node ast.Node) []MutatedNode {
	bin, ok := node.(*ast.BinaryExpr)
	if !ok {
		return nil
	}

	replacements, exists := logicalReplacements[bin.Op]
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