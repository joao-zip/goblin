package mutator

import (
	"go/ast"
	"go/token"
)

var assignmentReplacements = map[token.Token]token.Token{
	token.ADD_ASSIGN: token.SUB_ASSIGN, // += -> -=
	token.SUB_ASSIGN: token.ADD_ASSIGN, // -= -> +=
	token.MUL_ASSIGN: token.QUO_ASSIGN, // *= -> /=
	token.QUO_ASSIGN: token.MUL_ASSIGN, // /= -> *=
	token.REM_ASSIGN: token.MUL_ASSIGN, // %= -> *=
}

// AssignmentMutator swaps compound assignment operators: +=, -=, *=, /=, %=.
type AssignmentMutator struct{}

func (m *AssignmentMutator) Name() string { return "assignment" }

func (m *AssignmentMutator) CanMutate(node ast.Node) bool {
	assign, ok := node.(*ast.AssignStmt)
	if !ok {
		return false
	}
	_, exists := assignmentReplacements[assign.Tok]
	return exists
}

func (m *AssignmentMutator) Mutate(node ast.Node) []MutatedNode {
	assign, ok := node.(*ast.AssignStmt)
	if !ok {
		return nil
	}

	replacement, exists := assignmentReplacements[assign.Tok]
	if !exists {
		return nil
	}

	original := assign.Tok

	return []MutatedNode{
		{
			Original:    original.String(),
			Replacement: replacement.String(),
			Apply:       func() { assign.Tok = replacement },
			Rollback:    func() { assign.Tok = original },
		},
	}
}
