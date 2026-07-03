package mutator

// DefaultMutators returns all built-in mutators.
func DefaultMutators() []Mutator {
	return []Mutator{
		&ArithmeticMutator{},
		&ComparisonMutator{},
		&LogicalMutator{},
		&UnaryMutator{},
		&AssignmentMutator{},
	}
}

// FilterMutators filters mutators by name. If names is nil or empty, returns all.
func FilterMutators(mutators []Mutator, names []string) []Mutator {
	if len(names) == 0 {
		return mutators
	}

	allowed := make(map[string]bool, len(names))
	for _, n := range names {
		allowed[n] = true
	}

	var filtered []Mutator
	for _, m := range mutators {
		if allowed[m.Name()] {
			filtered = append(filtered, m)
		}
	}
	return filtered
}
