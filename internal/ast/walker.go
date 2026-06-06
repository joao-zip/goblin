package ast

import (
	"go/ast"
	"go/token"

	"github.com/joao-zip/gomutate/pkg/mutation"
)

// MutableNode represents an AST node that can be mutated.
type MutableNode struct {
	Node     ast.Node
	Type     mutation.MutationType
	File     string
	Line     int
	Column   int
	Original string
}

// arithmeticOps are binary operators classified as arithmetic mutations.
var arithmeticOps = map[token.Token]bool{
	token.ADD: true, // +
	token.SUB: true, // -
	token.MUL: true, // *
	token.QUO: true, // /
	token.REM: true, // %
}

// comparisonOps are binary operators classified as comparison mutations.
var comparisonOps = map[token.Token]bool{
	token.EQL: true, // ==
	token.NEQ: true, // !=
	token.LSS: true, // <
	token.GTR: true, // >
	token.LEQ: true, // <=
	token.GEQ: true, // >=
}

// logicalOps are binary operators classified as logical mutations.
var logicalOps = map[token.Token]bool{
	token.LAND: true, // &&
	token.LOR:  true, // ||
}

// assignOps are assignment operators that can be mutated.
var assignOps = map[token.Token]bool{
	token.ADD_ASSIGN: true, // +=
	token.SUB_ASSIGN: true, // -=
	token.MUL_ASSIGN: true, // *=
	token.QUO_ASSIGN: true, // /=
	token.REM_ASSIGN: true, // %=
}

// FindMutableNodes walks the AST and returns all nodes that can be mutated.
func FindMutableNodes(file *ast.File, fset *token.FileSet) []MutableNode {
	var nodes []MutableNode

	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		switch node := n.(type) {
		case *ast.BinaryExpr:
			if mt, ok := classifyBinaryOp(node.Op); ok {
				pos := fset.Position(node.OpPos)
				nodes = append(nodes, MutableNode{
					Node:     node,
					Type:     mt,
					File:     pos.Filename,
					Line:     pos.Line,
					Column:   pos.Column,
					Original: node.Op.String(),
				})
			}

		case *ast.AssignStmt:
			if assignOps[node.Tok] {
				pos := fset.Position(node.TokPos)
				nodes = append(nodes, MutableNode{
					Node:     node,
					Type:     mutation.Assignment,
					File:     pos.Filename,
					Line:     pos.Line,
					Column:   pos.Column,
					Original: node.Tok.String(),
				})
			}

		case *ast.IncDecStmt:
			pos := fset.Position(node.TokPos)
			nodes = append(nodes, MutableNode{
				Node:     node,
				Type:     mutation.Unary,
				File:     pos.Filename,
				Line:     pos.Line,
				Column:   pos.Column,
				Original: node.Tok.String(),
			})
		}

		return true
	})

	return nodes
}

func classifyBinaryOp(op token.Token) (mutation.MutationType, bool) {
	switch {
	case arithmeticOps[op]:
		return mutation.Arithmetic, true
	case comparisonOps[op]:
		return mutation.Comparison, true
	case logicalOps[op]:
		return mutation.Logical, true
	default:
		return "", false
	}
}
