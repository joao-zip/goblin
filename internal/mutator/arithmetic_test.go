package mutator

import (
	"go/ast"
	"go/token"
	"testing"
)

func TestMutatorInterface(t *testing.T) {
	// Verifica que ArithmeticMutator implementa a interface Mutator.
	var _ Mutator = &ArithmeticMutator{}
}

func TestArithmeticMutator_Name(t *testing.T) {
	m := &ArithmeticMutator{}
	if m.Name() != "arithmetic" {
		t.Errorf("Name() = %q, want %q", m.Name(), "arithmetic")
	}
}

func TestArithmeticMutator_CanMutate(t *testing.T) {
	m := &ArithmeticMutator{}

	tests := []struct {
		name string
		node ast.Node
		want bool
	}{
		{"add", &ast.BinaryExpr{Op: token.ADD}, true},
		{"sub", &ast.BinaryExpr{Op: token.SUB}, true},
		{"mul", &ast.BinaryExpr{Op: token.MUL}, true},
		{"quo", &ast.BinaryExpr{Op: token.QUO}, true},
		{"rem", &ast.BinaryExpr{Op: token.REM}, true},
		{"not arithmetic (==)", &ast.BinaryExpr{Op: token.EQL}, false},
		{"not arithmetic (&&)", &ast.BinaryExpr{Op: token.LAND}, false},
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

func TestArithmeticMutator_Mutate(t *testing.T) {
	m := &ArithmeticMutator{}

	tests := []struct {
		name     string
		op       token.Token
		wantOps  []token.Token
	}{
		{"add", token.ADD, []token.Token{token.SUB, token.MUL, token.QUO}},
		{"sub", token.SUB, []token.Token{token.ADD, token.MUL, token.QUO}},
		{"mul", token.MUL, []token.Token{token.ADD, token.SUB, token.QUO}},
		{"quo", token.QUO, []token.Token{token.ADD, token.SUB, token.MUL}},
		{"rem", token.REM, []token.Token{token.ADD, token.SUB, token.MUL}},
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

func TestArithmeticMutator_Mutate_NonBinaryExpr(t *testing.T) {
	m := &ArithmeticMutator{}
	mutations := m.Mutate(&ast.Ident{Name: "x"})
	if len(mutations) != 0 {
		t.Errorf("Mutate() on non-BinaryExpr returned %d mutations, want 0", len(mutations))
	}
}

func TestArithmeticMutator_ApplyAndRollback(t *testing.T) {
	m := &ArithmeticMutator{}
	node := &ast.BinaryExpr{Op: token.ADD}

	mutations := m.Mutate(node)
	if len(mutations) == 0 {
		t.Fatal("expected mutations")
	}

	mut := mutations[0] // + → -

	mut.Apply()
	if node.Op != token.SUB {
		t.Errorf("after Apply(), Op = %s, want %s", node.Op, token.SUB)
	}

	mut.Rollback()
	if node.Op != token.ADD {
		t.Errorf("after Rollback(), Op = %s, want %s", node.Op, token.ADD)
	}
}
