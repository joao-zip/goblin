package simple

import "testing"

func TestAdd(t *testing.T) {
	if Add(2, 3) != 5 {
		t.Error("Add(2, 3) should be 5")
	}
	if Add(-1, 1) != 0 {
		t.Error("Add(-1, 1) should be 0")
	}
}

func TestSub(t *testing.T) {
	if Sub(5, 3) != 2 {
		t.Error("Sub(5, 3) should be 2")
	}
}

func TestIsPositive(t *testing.T) {
	if !IsPositive(1) {
		t.Error("IsPositive(1) should be true")
	}
	if IsPositive(-1) {
		t.Error("IsPositive(-1) should be false")
	}
}
