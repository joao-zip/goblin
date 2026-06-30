package mutator

import (
	"go/ast"
	"go/token"
	"testing"
)

func TestAssignmentMutator_ImplementsMutator(t *testing.T) {
	var _ Mutator = &AssignmentMutator{}
}

func TestAssignmentMutator_Name(t *testing.T) {
	m := &AssignmentMutator{}
	if m.Name() != "assignment" {
		t.Errorf("Name() = %q, want %q", m.Name(), "assignment")
	}
}

func TestAssignmentMutator_CanMutate(t *testing.T) {
	m := &AssignmentMutator{}

	tests := []struct {
		name string
		node ast.Node
		want bool
	}{
		{"add_assign", &ast.AssignStmt{Tok: token.ADD_ASSIGN}, true},
		{"sub_assign", &ast.AssignStmt{Tok: token.SUB_ASSIGN}, true},
		{"mul_assign", &ast.AssignStmt{Tok: token.MUL_ASSIGN}, true},
		{"quo_assign", &ast.AssignStmt{Tok: token.QUO_ASSIGN}, true},
		{"rem_assign", &ast.AssignStmt{Tok: token.REM_ASSIGN}, true},
		{"not compound (assign)", &ast.AssignStmt{Tok: token.ASSIGN}, false},
		{"not assign stmt", &ast.Ident{Name: "x"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := m.CanMutate(tt.node); got != tt.want {
				t.Errorf("CanMutate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssignmentMutator_Mutate(t *testing.T) {
	m := &AssignmentMutator{}

	tests := []struct {
		name    string
		tok     token.Token
		wantRep token.Token
	}{
		{"add_assign to sub_assign", token.ADD_ASSIGN, token.SUB_ASSIGN},
		{"sub_assign to add_assign", token.SUB_ASSIGN, token.ADD_ASSIGN},
		{"mul_assign to quo_assign", token.MUL_ASSIGN, token.QUO_ASSIGN},
		{"quo_assign to mul_assign", token.QUO_ASSIGN, token.MUL_ASSIGN},
		{"rem_assign to mul_assign", token.REM_ASSIGN, token.MUL_ASSIGN},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &ast.AssignStmt{Tok: tt.tok}
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

func TestAssignmentMutator_Mutate_NonAssignStmt(t *testing.T) {
	m := &AssignmentMutator{}
	mutations := m.Mutate(&ast.Ident{Name: "x"})
	if len(mutations) != 0 {
		t.Errorf("Mutate() returned %d mutations for non-AssignStmt, want 0", len(mutations))
	}
}

func TestAssignmentMutator_ApplyAndRollback(t *testing.T) {
	m := &AssignmentMutator{}
	node := &ast.AssignStmt{Tok: token.ADD_ASSIGN}

	mutations := m.Mutate(node)
	if len(mutations) == 0 {
		t.Fatal("expected mutations")
	}

	mut := mutations[0]
	mut.Apply()
	if node.Tok != token.SUB_ASSIGN {
		t.Errorf("after Apply(), Tok = %s, want %s", node.Tok, token.SUB_ASSIGN)
	}

	mut.Rollback()
	if node.Tok != token.ADD_ASSIGN {
		t.Errorf("after Rollback(), Tok = %s, want %s", node.Tok, token.ADD_ASSIGN)
	}
}
