package main

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

	astutil "github.com/joao-zip/goblin/internal/ast"
	"github.com/joao-zip/goblin/internal/config"
	"github.com/joao-zip/goblin/internal/mutator"
	"github.com/joao-zip/goblin/internal/report"
	"github.com/joao-zip/goblin/internal/runner"
	"github.com/joao-zip/goblin/internal/version"
	"github.com/joao-zip/goblin/pkg/mutation"
)

const (
	colorReset = "\033[0m"
	colorBold  = "\033[1m"
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
	colorCyan  = "\033[36m"
)

type candidate struct {
	file        *ast.File
	filePath    string
	node        ast.Node
	mutatorName string
	mutatedNode mutator.MutatedNode
	line        int
	column      int
	original    string
}

func main() {
	cfg, err := config.FromArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError parsing CLI arguments: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	if cfg.ShowVersion {
		fmt.Print(version.Banner())
		os.Exit(0)
	}

	fmt.Printf("%s%sGoblin v%s — Mutation Testing for Go%s\n\n", colorBold, colorCyan, version.Version, colorReset)

	absDir, err := filepath.Abs(cfg.Dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError resolving directory path: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	fmt.Printf("Scanning directory: %s...\n", absDir)
	files, fset, err := astutil.ParseDir(absDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError parsing directory: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("No Go files found to mutate.")
		os.Exit(0)
	}

	muts := mutator.DefaultMutators()
	muts = mutator.FilterMutators(muts, cfg.Mutators)

	candidates := collectCandidates(files, fset, muts)

	totalMutants := len(candidates)
	if totalMutants == 0 {
		fmt.Println("No mutations could be generated.")
		os.Exit(0)
	}

	workers := cfg.Workers
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	fmt.Printf("Found %d mutation candidates in %d source files. (workers: %d)\n\n", totalMutants, len(files), workers)

	jobs, err := buildJobs(candidates, absDir, fset)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError preparing mutations: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	ctx := context.Background()
	results := runner.RunAll(ctx, jobs, runner.PoolConfig{
		Workers:    workers,
		Timeout:    runner.Config{Timeout: cfg.Timeout},
		ProjectDir: absDir,
	})

	for _, r := range results {
		printStatus(r)
	}

	textReporter := &report.TextReporter{}
	textReporter.Report(os.Stdout, results)

	if cfg.Output != "" {
		if err := writeJSONReport(cfg.Output, results); err != nil {
			fmt.Fprintf(os.Stderr, "%sError writing JSON report: %v%s\n", colorRed, err, colorReset)
			os.Exit(1)
		}
		fmt.Printf("\nJSON report written to: %s\n", cfg.Output)
	}

	score := report.CalculateScore(results)
	if cfg.Threshold > 0 {
		fmt.Printf("Threshold Required: %.2f%%\n", cfg.Threshold)
		if score < cfg.Threshold {
			fmt.Printf("\n%sFailure: Mutation score is below the required threshold of %.2f%%%s\n", colorRed, cfg.Threshold, colorReset)
			os.Exit(1)
		} else {
			fmt.Printf("\n%sSuccess: Mutation score meets or exceeds the threshold.%s\n", colorGreen, colorReset)
		}
	}
}

func collectCandidates(files []*ast.File, fset *token.FileSet, muts []mutator.Mutator) []candidate {
	var candidates []candidate
	for _, file := range files {
		nodes := astutil.FindMutableNodes(file, fset)
		for _, node := range nodes {
			for _, m := range muts {
				if m.CanMutate(node.Node) {
					mutated := m.Mutate(node.Node)
					for _, mn := range mutated {
						candidates = append(candidates, candidate{
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

func buildJobs(candidates []candidate, absDir string, fset *token.FileSet) ([]runner.Job, error) {
	jobs := make([]runner.Job, 0, len(candidates))

	for i, c := range candidates {
		originalData, err := os.ReadFile(c.filePath)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", c.filePath, err)
		}

		c.mutatedNode.Apply()
		var buf bytes.Buffer
		err = format.Node(&buf, fset, c.file)
		c.mutatedNode.Rollback()

		if err != nil {
			return nil, fmt.Errorf("formatting mutated AST for %s: %w", c.filePath, err)
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

func printStatus(r runner.Result) {
	const (
		green   = "\033[32m"
		red     = "\033[31m"
		yellow  = "\033[33m"
		magenta = "\033[35m"
		reset   = "\033[0m"
	)

	var label string
	switch r.Status {
	case mutation.Killed:
		label = green + "[KILLED]" + reset
	case mutation.Survived:
		label = red + "[SURVIVED]" + reset
	case mutation.Timeout:
		label = yellow + "[TIMEOUT]" + reset
	case mutation.Error:
		label = magenta + "[ERROR]" + reset
	}

	fmt.Printf(" [%d] %s:%d:%d (%s) %s → %s  %s\n",
		r.Mutation.ID, r.Mutation.File, r.Mutation.Line, r.Mutation.Column,
		r.Mutation.Type, r.Mutation.Original, r.Mutation.Replacement, label)
}

func writeJSONReport(path string, results []runner.Result) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating report file: %w", err)
	}
	defer f.Close()

	jsonReporter := &report.JSONReporter{}
	return jsonReporter.Report(f, results)
}
