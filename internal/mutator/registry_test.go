package mutator

import (
	"testing"
)

func TestDefaultMutators(t *testing.T) {
	mutators := DefaultMutators()

	expected := []string{"arithmetic", "comparison", "logical", "unary", "assignment"}

	if len(mutators) != len(expected) {
		t.Fatalf("DefaultMutators() returned %d mutators, want %d", len(mutators), len(expected))
	}

	for i, name := range expected {
		if mutators[i].Name() != name {
			t.Errorf("DefaultMutators()[%d].Name() = %q, want %q", i, mutators[i].Name(), name)
		}
	}
}

func TestFilterMutators_All(t *testing.T) {
	mutators := DefaultMutators()
	filtered := FilterMutators(mutators, []string{"arithmetic", "comparison", "logical", "unary", "assignment"})

	if len(filtered) != 5 {
		t.Fatalf("FilterMutators() returned %d, want 5", len(filtered))
	}
}

func TestFilterMutators_Subset(t *testing.T) {
	mutators := DefaultMutators()
	filtered := FilterMutators(mutators, []string{"arithmetic"})

	if len(filtered) != 1 {
		t.Fatalf("FilterMutators() returned %d, want 1", len(filtered))
	}
	if filtered[0].Name() != "arithmetic" {
		t.Errorf("filtered[0].Name() = %q, want %q", filtered[0].Name(), "arithmetic")
	}
}

func TestFilterMutators_Empty(t *testing.T) {
	mutators := DefaultMutators()
	filtered := FilterMutators(mutators, []string{})

	if len(filtered) != len(mutators) {
		t.Errorf("FilterMutators with empty list should return all, got %d want %d", len(filtered), len(mutators))
	}
}

func TestFilterMutators_Nil(t *testing.T) {
	mutators := DefaultMutators()
	filtered := FilterMutators(mutators, nil)

	if len(filtered) != len(mutators) {
		t.Errorf("FilterMutators with nil should return all, got %d want %d", len(filtered), len(mutators))
	}
}

func TestFilterMutators_Unknown(t *testing.T) {
	mutators := DefaultMutators()
	filtered := FilterMutators(mutators, []string{"nonexistent"})

	if len(filtered) != 0 {
		t.Errorf("FilterMutators with unknown name returned %d, want 0", len(filtered))
	}
}
