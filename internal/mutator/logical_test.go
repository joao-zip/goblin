package mutator

import (
	"go/ast"
	"go/token"
	"testing"
)

func TestLogicalMutator_Name(t *testing.T) {
	m := &LogicalMutator{}
	if m.Name() != "logical" {
		t.Errorf("Name() = %q, want %q", m.Name(), "logical")
	}
}

func TestLogicalMutator_CanMutate(t *testing.T) {
	m := &LogicalMutator{}

	tests := []struct {
		name string
		node ast.Node
		want bool
	}{
		{"land", &ast.BinaryExpr{Op: token.LAND}, true},
		{"lor", &ast.BinaryExpr{Op: token.LOR}, true},
		{"not logical (+)", &ast.BinaryExpr{Op: token.ADD}, false},
		{"not logical (==)", &ast.BinaryExpr{Op: token.EQL}, false},
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

func TestLogicalMutator_Mutate(t *testing.T) {
	m := &LogicalMutator{}

	tests := []struct {
		name    string
		op      token.Token
		wantOps []token.Token
	}{
		{"&& -> ||", token.LAND, []token.Token{token.LOR}},
		{"|| -> &&", token.LOR, []token.Token{token.LAND}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &ast.BinaryExpr{Op: tt.op}
			mutations := m.Mutate(node)

			if len(mutations) != len(tt.wantOps) {
				t.Fatalf("Mutate() returned %d mutations, want %d", len(mutations), len(tt.wantOps))
			}

			for i, mut := range mutations {
				if mut.Replacement != tt.wantOps[i].String() {
					t.Errorf("mutations[%d].Replacement = %q, want %q", i, mut.Replacement, tt.wantOps[i].String())
				}
			}
		})
	}
}

func TestLogicalMutator_Mutate_NonBinaryExpr(t *testing.T) {
	m := &LogicalMutator{}
	mutations := m.Mutate(&ast.Ident{Name: "x"})
	if len(mutations) != 0 {
		t.Errorf("Mutate() on non-BinaryExpr returned %d mutations, want 0", len(mutations))
	}
}

func TestLogicalMutator_ApplyAndRollback(t *testing.T) {
	m := &LogicalMutator{}
	node := &ast.BinaryExpr{Op: token.LAND}

	mutations := m.Mutate(node)
	if len(mutations) == 0 {
		t.Fatal("expected mutations")
	}

	mut := mutations[0] // && -> ||

	mut.Apply()
	if node.Op != token.LOR {
		t.Errorf("after Apply(), Op = %s, want %s", node.Op, token.LOR)
	}

	mut.Rollback()
	if node.Op != token.LAND {
		t.Errorf("after Rollback(), Op = %s, want %s", node.Op, token.LAND)
	}
}
