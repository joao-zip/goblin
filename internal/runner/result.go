package runner

import (
	"time"

	"github.com/joao-zip/gomutate/pkg/mutation"
)

// Result holds the outcome of running tests against a single mutant.
type Result struct {
	Mutation mutation.Mutation
	Status   mutation.MutationStatus
	Output   string
	Duration time.Duration
}
