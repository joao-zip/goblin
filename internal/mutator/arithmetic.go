package mutator

import (
	"go/ast"
	"go/token"
)

var arithmeticReplacements = map[token.Token][]token.Token{
	token.ADD: {token.SUB}, // + -> -
	token.SUB: {token.ADD}, // - -> +
	token.MUL: {token.QUO}, // * -> /
	token.QUO: {token.MUL}, // / -> *
	token.REM: {token.MUL}, // % -> *
}

// ArithmeticMutator swaps arithmetic operators: +, -, *, /, %.
type ArithmeticMutator struct{}

func (m *ArithmeticMutator) Name() string { return "arithmetic" }

func (m *ArithmeticMutator) CanMutate(node ast.Node) bool {
	bin, ok := node.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	_, exists := arithmeticReplacements[bin.Op]
	return exists
}

func (m *ArithmeticMutator) Mutate(node ast.Node) []MutatedNode {
	bin, ok := node.(*ast.BinaryExpr)
	if !ok {
		return nil
	}

	replacements, exists := arithmeticReplacements[bin.Op]
	if !exists {
		return nil
	}

	original := bin.Op
	var mutations []MutatedNode

	for _, rep := range replacements {
		rep := rep // capture loop variable
		mutations = append(mutations, MutatedNode{
			Original:    original.String(),
			Replacement: rep.String(),
			Apply:       func() { bin.Op = rep },
			Rollback:    func() { bin.Op = original },
		})
	}

	return mutations
}
