package mutator

import (
	"go/ast"
	"go/token"
	"testing"
)

func TestUnaryMutator_ImplementsMutator(t *testing.T) {
	var _ Mutator = &UnaryMutator{}
}

func TestUnaryMutator_Name(t *testing.T) {
	m := &UnaryMutator{}
	if m.Name() != "unary" {
		t.Errorf("Name() = %q, want %q", m.Name(), "unary")
	}
}

func TestUnaryMutator_CanMutate(t *testing.T) {
	m := &UnaryMutator{}

	tests := []struct {
		name string
		node ast.Node
		want bool
	}{
		{"inc", &ast.IncDecStmt{Tok: token.INC}, true},
		{"dec", &ast.IncDecStmt{Tok: token.DEC}, true},
		{"not unary (add)", &ast.IncDecStmt{Tok: token.ADD}, false},
		{"not incdecstmt", &ast.Ident{Name: "x"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := m.CanMutate(tt.node); got != tt.want {
				t.Errorf("CanMutate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnaryMutator_Mutate(t *testing.T) {
	m := &UnaryMutator{}

	tests := []struct {
		name    string
		tok     token.Token
		wantRep token.Token
	}{
		{"inc to dec", token.INC, token.DEC},
		{"dec to inc", token.DEC, token.INC},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &ast.IncDecStmt{Tok: tt.tok}
			mutations := m.Mutate(node)

			if len(mutations) != 1 {
				t.Fatalf("Mutate() returned %d mutations, want 1", len(mutations))
			}

			mut := mutations[0]
			if mut.Original != tt.tok.String() {
				t.Errorf("mut.Original = %q, want %q", mut.Original, tt.tok.String())
			}
			if mut.Replacement != tt.wantRep.String() {
				t.Errorf("mut.Replacement = %q, want %q", mut.Replacement, tt.wantRep.String())
			}
		})
	}
}

func TestUnaryMutator_Mutate_NonIncDecStmt(t *testing.T) {
	m := &UnaryMutator{}
	mutations := m.Mutate(&ast.Ident{Name: "x"})
	if len(mutations) != 0 {
		t.Errorf("Mutate() returned %d mutations for non-IncDecStmt, want 0", len(mutations))
	}
}

func TestUnaryMutator_ApplyAndRollback(t *testing.T) {
	m := &UnaryMutator{}
	node := &ast.IncDecStmt{Tok: token.INC}

	mutations := m.Mutate(node)
	if len(mutations) == 0 {
		t.Fatal("expected mutations")
	}

	mut := mutations[0]
	mut.Apply()
	if node.Tok != token.DEC {
		t.Errorf("after Apply(), Tok = %s, want %s", node.Tok, token.DEC)
	}

	mut.Rollback()
	if node.Tok != token.INC {
		t.Errorf("after Rollback(), Tok = %s, want %s", node.Tok, token.INC)
	}
}
