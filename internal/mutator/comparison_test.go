package mutator

import (
	"go/ast"
	"go/token"
	"testing"
)

func TestComparisonMutator_Name(t *testing.T) {
	m := &ComparisonMutator{}
	if m.Name() != "comparison" {
		t.Errorf("Name() = %q, want %q", m.Name(), "comparison")
	}
}

func TestComparisonMutator_CanMutate(t *testing.T) {
	m := &ComparisonMutator{}

	tests := []struct {
		name string
		node ast.Node
		want bool
	}{
		{"EQL (==)", &ast.BinaryExpr{Op: token.EQL}, true},
		{"NEQ (!=)", &ast.BinaryExpr{Op: token.NEQ}, true},
		{"LSS (<)", &ast.BinaryExpr{Op: token.LSS}, true},
		{"GEQ (>=)", &ast.BinaryExpr{Op: token.GEQ}, true},
		{"not comparison (>)", &ast.BinaryExpr{Op: token.GTR}, false},
		{"not comparison (<=)", &ast.BinaryExpr{Op: token.LEQ}, false},
		{"not comparison (+)", &ast.BinaryExpr{Op: token.ADD}, false},
		{"not comparison (&&)", &ast.BinaryExpr{Op: token.LAND}, false},
		{"not binary expr", &ast.Ident{Name: "x"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := m.CanMutate(tt.node); got != tt.want {
				t.Errorf("CanMutate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComparisonMutator_Mutate(t *testing.T) {
	m := &ComparisonMutator{}

	tests := []struct {
		name    string
		op      token.Token
		wantOps []token.Token
	}{
		{"EQL (==)", token.EQL, []token.Token{token.NEQ}},
		{"NEQ (!=)", token.NEQ, []token.Token{token.EQL}},
		{"LSS (<)", token.LSS, []token.Token{token.GEQ}},
		{"GEQ (>=)", token.GEQ, []token.Token{token.LSS}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &ast.BinaryExpr{Op: tt.op}
			mutations := m.Mutate(node)

			if len(mutations) != len(tt.wantOps) {
				t.Fatalf("Mutate() returned %d mutations, want %d", len(mutations), len(tt.wantOps))
			}

			for i, mut := range mutations {
				if mut.Original != tt.op.String() {
					t.Errorf("mutations[%d].Original = %q, want %q", i, mut.Original, tt.op.String())
				}
				if mut.Replacement != tt.wantOps[i].String() {
					t.Errorf("mutations[%d].Replacement = %q, want %q", i, mut.Replacement, tt.wantOps[i].String())
				}
			}
		})
	}
}

func TestComparisonMutator_Mutate_NonBinaryExpr(t *testing.T) {
	m := &ComparisonMutator{}
	mutations := m.Mutate(&ast.Ident{Name: "x"})
	if len(mutations) != 0 {
		t.Errorf("Mutate() on non-BinaryExpr returned %d mutations, want 0", len(mutations))
	}
}

func TestComparisonMutator_ApplyAndRollback(t *testing.T) {
	m := &ComparisonMutator{}
	node := &ast.BinaryExpr{Op: token.EQL}

	mutations := m.Mutate(node)
	if len(mutations) == 0 {
		t.Fatal("expected mutations")
	}

	mut := mutations[0] // == -> !=

	mut.Apply()
	if node.Op != token.NEQ {
		t.Errorf("after Apply(), Op = %s, want %s", node.Op, token.NEQ)
	}

	mut.Rollback()
	if node.Op != token.EQL {
		t.Errorf("after Rollback(), Op = %s, want %s", node.Op, token.EQL)
	}
}
