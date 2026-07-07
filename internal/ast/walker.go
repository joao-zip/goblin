package ast

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/joao-zip/goblin/pkg/mutation"
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

const ignoreComment = "goblin:ignore"

// findIgnoredLines returns a set of line numbers that should be skipped during mutation.
// A line is ignored if it contains a "// goblin:ignore" comment directly on it,
// or if the line immediately above it has a standalone "// goblin:ignore" comment.
func findIgnoredLines(file *ast.File, fset *token.FileSet) map[int]bool {
	ignored := make(map[int]bool)

	for _, cg := range file.Comments {
		for _, c := range cg.List {
			text := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
			if strings.TrimSpace(text) == ignoreComment {
				commentLine := fset.Position(c.Slash).Line
				// Always ignore the line the comment is on (covers inline comments).
				ignored[commentLine] = true
				// Only mark the next line when this comment is standalone
				// (i.e. it is the first comment in its group and the group starts
				// on the same line — meaning nothing else precedes it on that line).
				// We detect standalone by checking that the comment group starts
				// at column 1 or that the comment is the sole element on its line.
				commentCol := fset.Position(c.Slash).Column
				if commentCol <= 2 { // "//" at column 1 means standalone
					ignored[commentLine+1] = true
				}
			}
		}
	}

	return ignored
}

// FindMutableNodes walks the AST and returns all nodes that can be mutated,
// excluding any nodes on lines marked with // goblin:ignore.
func FindMutableNodes(file *ast.File, fset *token.FileSet) []MutableNode {
	var nodes []MutableNode
	ignored := findIgnoredLines(file, fset)

	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		switch node := n.(type) {
		case *ast.BinaryExpr:
			if mt, ok := classifyBinaryOp(node.Op); ok {
				pos := fset.Position(node.OpPos)
				if ignored[pos.Line] {
					return true
				}
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
				if ignored[pos.Line] {
					return true
				}
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
			if ignored[pos.Line] {
				return true
			}
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

