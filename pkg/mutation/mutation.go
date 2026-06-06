// Package mutation defines the core domain types for GoMutate.
package mutation

import "fmt"

// MutationType identifies the category of a mutation.
type MutationType string

const (
	Arithmetic MutationType = "arithmetic"
	Comparison MutationType = "comparison"
	Logical    MutationType = "logical"
	Unary      MutationType = "unary"
	Assignment MutationType = "assignment"
	Branch     MutationType = "branch"
	Literal    MutationType = "literal"
)

// AllMutationTypes returns all supported mutation types.
func AllMutationTypes() []MutationType {
	return []MutationType{
		Arithmetic, Comparison, Logical, Unary, Assignment, Branch, Literal,
	}
}

// MutationStatus represents the outcome of running tests against a mutant.
type MutationStatus string

const (
	Pending  MutationStatus = "pending"
	Killed   MutationStatus = "killed"
	Survived MutationStatus = "survived"
	Timeout  MutationStatus = "timeout"
	Error    MutationStatus = "error"
)

// IsTerminal reports whether the status is a final result (not pending).
func (s MutationStatus) IsTerminal() bool {
	return s != Pending
}

// AllMutationStatuses returns all possible mutation statuses.
func AllMutationStatuses() []MutationStatus {
	return []MutationStatus{
		Pending, Killed, Survived, Timeout, Error,
	}
}

// Mutation represents a single mutation applied to a source file.
type Mutation struct {
	ID          int
	Type        MutationType
	File        string
	Line        int
	Column      int
	Original    string
	Replacement string
	Status      MutationStatus
}

// String implements fmt.Stringer.
func (m Mutation) String() string {
	return fmt.Sprintf("[%s] %s:%d:%d — %s → %s (%s)",
		m.Type, m.File, m.Line, m.Column, m.Original, m.Replacement, m.Status,
	)
}
