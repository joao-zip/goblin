package demo

import "testing"

// "Testes" que parecem bons mas são fracos:
// - Não testam limites
// - Não verificam valores negativos
// - Alguns nem verificam o resultado

func TestIsAdult(t *testing.T) {
	if !IsAdult(25) {
		t.Error("25 should be adult")
	}
}

func TestDiscount(t *testing.T) {
	result := Discount(100, 10)
	if result != 90 {
		t.Errorf("expected 90, got %f", result)
	}
}

func TestMax(t *testing.T) {
	if Max(5, 3) != 5 {
		t.Error("Max(5,3) should be 5")
	}
}
