// Package goblin provides a programmatic API for running mutation tests on Go projects.
//
// Use [Run] to execute mutation testing with custom options:
//
//	result, err := goblin.Run(goblin.Options{
//	    Dir:     "./my-project",
//	    Timeout: 30 * time.Second,
//	})
//	fmt.Printf("Score: %.2f%%\n", result.Score)
package goblin

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"time"

	astutil "github.com/joao-zip/goblin/internal/ast"
	"github.com/joao-zip/goblin/internal/mutator"
	"github.com/joao-zip/goblin/internal/report"
	"github.com/joao-zip/goblin/internal/runner"
	"github.com/joao-zip/goblin/internal/version"
)

// Version is the current version of the Goblin library.
var Version = version.Version

// Options configures a mutation testing run.
type Options struct {
	// Dir is the project directory to test. Defaults to ".".
	Dir string

	// Timeout is the time limit for each test execution. Defaults to 10s.
	Timeout time.Duration

	// Mutators is a list of mutator names to activate.
	// If nil or empty, all built-in mutators are used.
	// Valid names: "arithmetic", "comparison", "logical", "unary", "assignment".
	Mutators []string

	// Workers is the number of parallel workers. Defaults to runtime.NumCPU().
	Workers int

	// Threshold is the minimum mutation score percentage.
	// If the score is below this value, Run returns an error.
	// Set to 0 to disable threshold checking.
	Threshold float64

	// HTML is the output file path for the interactive HTML report.
	// If empty, no HTML report is written.
	HTML string
}

// Result contains the complete results of a mutation testing run.
type Result struct {
	// Mutants contains the individual result for each mutant.
	Mutants []MutantResult

	// Score is the mutation score as a percentage (0–100).
	Score float64

	// Summary contains aggregate counts.
	Summary Summary
}

// Summary contains aggregate mutation testing statistics.
type Summary struct {
	Total    int
	Killed   int
	Survived int
	Timeout  int
	Errors   int
}

// MutantResult contains the result of a single mutation.
type MutantResult struct {
	ID          int
	Type        string
	File        string
	Line        int
	Column      int
	Original    string
	Replacement string
	Status      string // "killed", "survived", "timeout", "error"
	Duration    time.Duration
}

// ThresholdError is returned when the mutation score is below the configured threshold.
type ThresholdError struct {
	Score     float64
	Threshold float64
}

func (e *ThresholdError) Error() string {
	return "mutation score " + formatFloat(e.Score) + "% is below threshold " + formatFloat(e.Threshold) + "%"
}

func formatFloat(f float64) string {
	s := ""
	whole := int(f)
	frac := int((f - float64(whole)) * 100)
	if frac < 0 {
		frac = -frac
	}
	s += itoa(whole) + "." + pad2(frac)
	return s
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}

func pad2(n int) string {
	if n < 10 {
		return "0" + itoa(n)
	}
	return itoa(n)
}

// Run executes mutation testing on the specified directory with the given options.
// It returns structured results suitable for programmatic consumption.
func Run(opts Options) (*Result, error) {
	if opts.Dir == "" {
		opts.Dir = "."
	}
	if opts.Timeout == 0 {
		opts.Timeout = 10 * time.Second
	}
	if opts.Workers <= 0 {
		opts.Workers = runtime.NumCPU()
	}

	absDir, err := filepath.Abs(opts.Dir)
	if err != nil {
		return nil, err
	}

	files, fset, err := astutil.ParseDir(absDir)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return &Result{}, nil
	}

	muts := mutator.DefaultMutators()
	muts = mutator.FilterMutators(muts, opts.Mutators)

	candidates := collectCandidates(files, fset, muts)
	if len(candidates) == 0 {
		return &Result{}, nil
	}

	jobs, err := buildJobs(candidates, absDir, fset)
	if err != nil {
		return nil, err
	}

	workers := opts.Workers
	if workers > len(jobs) {
		workers = len(jobs)
	}

	ctx := context.Background()
	runnerResults := runner.RunAll(ctx, jobs, runner.PoolConfig{
		Workers:    workers,
		Timeout:    runner.Config{Timeout: opts.Timeout},
		ProjectDir: absDir,
	})

	score := report.CalculateScore(runnerResults)

	result := &Result{
		Score: score,
	}

	var killed, survived, timeout, errors int
	for _, r := range runnerResults {
		mr := MutantResult{
			ID:          r.Mutation.ID,
			Type:        string(r.Mutation.Type),
			File:        r.Mutation.File,
			Line:        r.Mutation.Line,
			Column:      r.Mutation.Column,
			Original:    r.Mutation.Original,
			Replacement: r.Mutation.Replacement,
			Status:      string(r.Status),
			Duration:    r.Duration,
		}
		result.Mutants = append(result.Mutants, mr)

		switch string(r.Status) {
		case "killed":
			killed++
		case "survived":
			survived++
		case "timeout":
			timeout++
		case "error":
			errors++
		}
	}

	result.Summary = Summary{
		Total:    len(runnerResults),
		Killed:   killed,
		Survived: survived,
		Timeout:  timeout,
		Errors:   errors,
	}

	if opts.HTML != "" {
		f, err := os.Create(opts.HTML)
		if err != nil {
			return result, fmt.Errorf("creating HTML report: %w", err)
		}
		defer f.Close()
		if err := (&report.HTMLReporter{}).Report(f, runnerResults); err != nil {
			return result, fmt.Errorf("writing HTML report: %w", err)
		}
	}

	if opts.Threshold > 0 && score < opts.Threshold {
		return result, &ThresholdError{Score: score, Threshold: opts.Threshold}
	}

	return result, nil
}

type candidateEntry struct {
	file        *ast.File
	filePath    string
	node        ast.Node
	mutatorName string
	mutatedNode mutator.MutatedNode
	line        int
	column      int
	original    string
}

func collectCandidates(files []*ast.File, fset *token.FileSet, muts []mutator.Mutator) []candidateEntry {
	var candidates []candidateEntry
	for _, file := range files {
		nodes := astutil.FindMutableNodes(file, fset)
		for _, node := range nodes {
			for _, m := range muts {
				if m.CanMutate(node.Node) {
					mutated := m.Mutate(node.Node)
					for _, mn := range mutated {
						candidates = append(candidates, candidateEntry{
							file:        file,
							filePath:    node.File,
							node:        node.Node,
							mutatorName: m.Name(),
							mutatedNode: mn,
							line:        node.Line,
							column:      node.Column,
							original:    node.Original,
						})
					}
				}
			}
		}
	}
	return candidates
}

func buildJobs(candidates []candidateEntry, absDir string, fset *token.FileSet) ([]runner.Job, error) {
	jobs := make([]runner.Job, 0, len(candidates))

	for i, c := range candidates {
		originalData, err := os.ReadFile(c.filePath)
		if err != nil {
			return nil, err
		}

		c.mutatedNode.Apply()
		var buf bytes.Buffer
		err = format.Node(&buf, fset, c.file)
		c.mutatedNode.Rollback()

		if err != nil {
			return nil, err
		}

		relPath, _ := filepath.Rel(absDir, c.filePath)
		if relPath == "" {
			relPath = filepath.Base(c.filePath)
		}

		jobs = append(jobs, runner.Job{
			ID:           i + 1,
			RelPath:      relPath,
			OriginalData: originalData,
			MutatedData:  buf.Bytes(),
			MutatorName:  c.mutatorName,
			Line:         c.line,
			Column:       c.column,
			Original:     c.original,
			Replacement:  c.mutatedNode.Replacement,
		})
	}

	return jobs, nil
}
