package mutation

import (
	"strings"
	"testing"
)

func TestMutationTypeValues(t *testing.T) {
	tests := []struct {
		name     string
		got      MutationType
		expected string
	}{
		{"arithmetic", Arithmetic, "arithmetic"},
		{"comparison", Comparison, "comparison"},
		{"logical", Logical, "logical"},
		{"unary", Unary, "unary"},
		{"assignment", Assignment, "assignment"},
		{"branch", Branch, "branch"},
		{"literal", Literal, "literal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.expected {
				t.Errorf("MutationType %s = %q, want %q", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestMutationStatusValues(t *testing.T) {
	tests := []struct {
		name     string
		got      MutationStatus
		expected string
	}{
		{"pending", Pending, "pending"},
		{"killed", Killed, "killed"},
		{"survived", Survived, "survived"},
		{"timeout", Timeout, "timeout"},
		{"error", Error, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.expected {
				t.Errorf("MutationStatus %s = %q, want %q", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestMutationStatusIsTerminal(t *testing.T) {
	tests := []struct {
		status   MutationStatus
		terminal bool
	}{
		{Pending, false},
		{Killed, true},
		{Survived, true},
		{Timeout, true},
		{Error, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsTerminal(); got != tt.terminal {
				t.Errorf("%s.IsTerminal() = %v, want %v", tt.status, got, tt.terminal)
			}
		})
	}
}

func TestMutationString(t *testing.T) {
	m := Mutation{
		ID:          1,
		Type:        Arithmetic,
		File:        "calc.go",
		Line:        10,
		Column:      5,
		Original:    "+",
		Replacement: "-",
		Status:      Pending,
	}

	s := m.String()

	checks := []struct {
		label    string
		contains string
	}{
		{"type", "arithmetic"},
		{"file", "calc.go"},
		{"line", "10"},
		{"original", "+"},
		{"replacement", "-"},
	}

	for _, c := range checks {
		t.Run(c.label, func(t *testing.T) {
			if !strings.Contains(s, c.contains) {
				t.Errorf("String() = %q, should contain %q", s, c.contains)
			}
		})
	}
}

func TestMutationZeroValue(t *testing.T) {
	var m Mutation

	if m.ID != 0 {
		t.Errorf("zero value ID = %d, want 0", m.ID)
	}
	if m.Type != "" {
		t.Errorf("zero value Type = %q, want empty", m.Type)
	}
	if m.Status != "" {
		t.Errorf("zero value Status = %q, want empty", m.Status)
	}
	if m.File != "" {
		t.Errorf("zero value File = %q, want empty", m.File)
	}
}

func TestAllMutationTypes(t *testing.T) {
	types := AllMutationTypes()

	expected := []MutationType{Arithmetic, Comparison, Logical, Unary, Assignment, Branch, Literal}

	if len(types) != len(expected) {
		t.Fatalf("AllMutationTypes() returned %d types, want %d", len(types), len(expected))
	}

	for i, mt := range expected {
		if types[i] != mt {
			t.Errorf("AllMutationTypes()[%d] = %q, want %q", i, types[i], mt)
		}
	}
}

func TestAllMutationStatuses(t *testing.T) {
	statuses := AllMutationStatuses()

	expected := []MutationStatus{Pending, Killed, Survived, Timeout, Error}

	if len(statuses) != len(expected) {
		t.Fatalf("AllMutationStatuses() returned %d statuses, want %d", len(statuses), len(expected))
	}

	for i, ms := range expected {
		if statuses[i] != ms {
			t.Errorf("AllMutationStatuses()[%d] = %q, want %q", i, statuses[i], ms)
		}
	}
}
