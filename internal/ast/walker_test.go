package ast

import (
	"testing"

	"github.com/joao-zip/goblin/pkg/mutation"
)

func TestFindMutableNodes_Arithmetic(t *testing.T) {
	src := `package example

func calc(a, b int) int {
	return a + b
}
`
	file, fset, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource() error = %v", err)
	}

	nodes := FindMutableNodes(file, fset)

	found := filterByType(nodes, mutation.Arithmetic)
	if len(found) != 1 {
		t.Fatalf("got %d arithmetic nodes, want 1", len(found))
	}
	if found[0].Original != "+" {
		t.Errorf("Original = %q, want %q", found[0].Original, "+")
	}
}

func TestFindMutableNodes_MultipleArithmetic(t *testing.T) {
	src := `package example

func calc(a, b, c int) int {
	return a + b - c * 2
}
`
	file, fset, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource() error = %v", err)
	}

	nodes := FindMutableNodes(file, fset)

	found := filterByType(nodes, mutation.Arithmetic)
	if len(found) != 3 {
		t.Fatalf("got %d arithmetic nodes, want 3", len(found))
	}
}

func TestFindMutableNodes_Comparison(t *testing.T) {
	src := `package example

func check(a, b int) bool {
	return a > b
}
`
	file, fset, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource() error = %v", err)
	}

	nodes := FindMutableNodes(file, fset)

	found := filterByType(nodes, mutation.Comparison)
	if len(found) != 1 {
		t.Fatalf("got %d comparison nodes, want 1", len(found))
	}
	if found[0].Original != ">" {
		t.Errorf("Original = %q, want %q", found[0].Original, ">")
	}
}

func TestFindMutableNodes_Logical(t *testing.T) {
	src := `package example

func check(a, b bool) bool {
	return a && b
}
`
	file, fset, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource() error = %v", err)
	}

	nodes := FindMutableNodes(file, fset)

	found := filterByType(nodes, mutation.Logical)
	if len(found) != 1 {
		t.Fatalf("got %d logical nodes, want 1", len(found))
	}
	if found[0].Original != "&&" {
		t.Errorf("Original = %q, want %q", found[0].Original, "&&")
	}
}

func TestFindMutableNodes_MixedExpression(t *testing.T) {
	src := `package example

func validate(x, y int) bool {
	return x > 0 && y < 10
}
`
	file, fset, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource() error = %v", err)
	}

	nodes := FindMutableNodes(file, fset)

	comparisons := filterByType(nodes, mutation.Comparison)
	logicals := filterByType(nodes, mutation.Logical)

	if len(comparisons) != 2 {
		t.Errorf("got %d comparison nodes, want 2", len(comparisons))
	}
	if len(logicals) != 1 {
		t.Errorf("got %d logical nodes, want 1", len(logicals))
	}
}

func TestFindMutableNodes_NoMutableNodes(t *testing.T) {
	src := `package example

func hello() string {
	return "hello"
}
`
	file, fset, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource() error = %v", err)
	}

	nodes := FindMutableNodes(file, fset)
	if len(nodes) != 0 {
		t.Errorf("got %d nodes, want 0", len(nodes))
	}
}

func TestFindMutableNodes_PositionInfo(t *testing.T) {
	src := `package example

func add(a, b int) int {
	return a + b
}
`
	file, fset, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource() error = %v", err)
	}

	nodes := FindMutableNodes(file, fset)
	if len(nodes) == 0 {
		t.Fatal("expected at least one node")
	}

	n := nodes[0]
	if n.Line <= 0 {
		t.Errorf("Line = %d, want > 0", n.Line)
	}
	if n.Column <= 0 {
		t.Errorf("Column = %d, want > 0", n.Column)
	}
	if n.File == "" {
		t.Error("File should not be empty")
	}
}

func TestFindMutableNodes_Assignment(t *testing.T) {
	src := `package example

func inc(a int) int {
	a += 1
	return a
}
`
	file, fset, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource() error = %v", err)
	}

	nodes := FindMutableNodes(file, fset)

	found := filterByType(nodes, mutation.Assignment)
	if len(found) != 1 {
		t.Fatalf("got %d assignment nodes, want 1", len(found))
	}
	if found[0].Original != "+=" {
		t.Errorf("Original = %q, want %q", found[0].Original, "+=")
	}
}

func TestFindMutableNodes_IncDec(t *testing.T) {
	src := `package example

func inc(a int) int {
	a++
	return a
}
`
	file, fset, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource() error = %v", err)
	}

	nodes := FindMutableNodes(file, fset)

	found := filterByType(nodes, mutation.Unary)
	if len(found) != 1 {
		t.Fatalf("got %d unary nodes, want 1", len(found))
	}
	if found[0].Original != "++" {
		t.Errorf("Original = %q, want %q", found[0].Original, "++")
	}
}

// filterByType is a test helper that filters MutableNodes by MutationType.
func filterByType(nodes []MutableNode, mt mutation.MutationType) []MutableNode {
	var result []MutableNode
	for _, n := range nodes {
		if n.Type == mt {
			result = append(result, n)
		}
	}
	return result
}

func TestFindMutableNodes_IgnoreInlineComment(t *testing.T) {
	src := `package example

func calc(a, b int) int {
	return a + b // goblin:ignore
}
`
	file, fset, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource() error = %v", err)
	}

	nodes := FindMutableNodes(file, fset)
	if len(nodes) != 0 {
		t.Errorf("expected 0 nodes (line ignored), got %d", len(nodes))
	}
}

func TestFindMutableNodes_IgnoreStandaloneComment(t *testing.T) {
	src := `package example

func calc(a, b int) int {
	// goblin:ignore
	return a + b
}
`
	file, fset, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource() error = %v", err)
	}

	nodes := FindMutableNodes(file, fset)
	if len(nodes) != 0 {
		t.Errorf("expected 0 nodes (next line ignored), got %d", len(nodes))
	}
}

func TestFindMutableNodes_IgnoreOnlyTaggedLines(t *testing.T) {
	src := `package example

func calc(a, b, c int) int {
	x := a + b // goblin:ignore
	return x - c
}
`
	file, fset, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource() error = %v", err)
	}

	nodes := FindMutableNodes(file, fset)
	// Only the subtraction on the return line should survive
	if len(nodes) != 1 {
		t.Errorf("expected 1 node (subtraction), got %d", len(nodes))
	}
	if len(nodes) == 1 && nodes[0].Original != "-" {
		t.Errorf("expected '-' operator, got %q", nodes[0].Original)
	}
}
