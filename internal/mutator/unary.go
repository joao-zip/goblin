package mutator

import (
	"go/ast"
	"go/token"
)

// UnaryMutator swaps ++ and --.
type UnaryMutator struct{}

func (m *UnaryMutator) Name() string { return "unary" }

func (m *UnaryMutator) CanMutate(node ast.Node) bool {
	stmt, ok := node.(*ast.IncDecStmt)
	if !ok {
		return false
	}
	return stmt.Tok == token.INC || stmt.Tok == token.DEC
}

func (m *UnaryMutator) Mutate(node ast.Node) []MutatedNode {
	stmt, ok := node.(*ast.IncDecStmt)
	if !ok {
		return nil
	}

	original := stmt.Tok
	var replacement token.Token
	if original == token.INC {
		replacement = token.DEC
	} else if original == token.DEC {
		replacement = token.INC
	} else {
		return nil
	}

	return []MutatedNode{
		{
			Original:    original.String(),
			Replacement: replacement.String(),
			Apply:       func() { stmt.Tok = replacement },
			Rollback:    func() { stmt.Tok = original },
		},
	}
}
