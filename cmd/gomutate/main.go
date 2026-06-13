package main

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"time"

	"github.com/joao-zip/gomutate/internal/config"
	astutil "github.com/joao-zip/gomutate/internal/ast"
	"github.com/joao-zip/gomutate/internal/mutator"
	"github.com/joao-zip/gomutate/internal/runner"
	"github.com/joao-zip/gomutate/pkg/mutation"
)

// ANSI color escape codes
const (
	colorReset   = "\033[0m"
	colorBold    = "\033[1m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
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
	fmt.Printf("%s%s🧬 GoMutate v0.1.0 — Mutation Testing for Go%s\n\n", colorBold, colorCyan, colorReset)

	cfg, err := config.FromArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError parsing CLI arguments: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

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

	totalMutants := len(candidates)
	fmt.Printf("Found %d mutation candidates in %d source files.\n\n", totalMutants, len(files))

	if totalMutants == 0 {
		fmt.Println("No mutations could be generated.")
		os.Exit(0)
	}

	var results []runner.Result
	var killed, survived, timeout, errCount int

	ctx := context.Background()

	for i, c := range candidates {
		id := i + 1
		relPath, _ := filepath.Rel(absDir, c.filePath)
		if relPath == "" {
			relPath = filepath.Base(c.filePath)
		}

		fmt.Printf(" [%d/%d] Mutating %s:%d:%d (%s) | %s → %s... ",
			id, totalMutants, relPath, c.line, c.column, c.mutatorName, c.original, c.mutatedNode.Replacement)

		res, runErr := runCandidate(ctx, c, absDir, cfg.Timeout, fset)
		if runErr != nil {
			fmt.Printf("%s[ERROR: %v]%s\n", colorRed, runErr, colorReset)
			errCount++
			results = append(results, runner.Result{
				Mutation: mutation.Mutation{
					ID:     id,
					Status: mutation.Error,
				},
				Status: mutation.Error,
			})
			continue
		}

		res.Mutation.ID = id
		results = append(results, res)

		switch res.Status {
		case mutation.Killed:
			fmt.Printf("%s[KILLED]%s\n", colorGreen, colorReset)
			killed++
		case mutation.Survived:
			fmt.Printf("%s[SURVIVED]%s\n", colorRed, colorReset)
			survived++
		case mutation.Timeout:
			fmt.Printf("%s[TIMEOUT]%s\n", colorYellow, colorReset)
			timeout++
		case mutation.Error:
			fmt.Printf("%s[BUILD ERROR]%s\n", colorMagenta, colorReset)
			errCount++
		}
	}

	fmt.Printf("\n%s--- Mutation Testing Results ---%s\n", colorBold, colorReset)
	fmt.Printf("Total Mutants:  %d\n", totalMutants)
	fmt.Printf("  %s✅ Killed:%s      %d (%.1f%%)\n", colorGreen, colorReset, killed, float64(killed)/float64(totalMutants)*100)
	fmt.Printf("  %s❌ Survived:%s    %d (%.1f%%)\n", colorRed, colorReset, survived, float64(survived)/float64(totalMutants)*100)
	fmt.Printf("  %s⏰ Timeout:%s     %d (%.1f%%)\n", colorYellow, colorReset, timeout, float64(timeout)/float64(totalMutants)*100)
	fmt.Printf("  %s⚠️  Error:%s       %d (%.1f%%)\n", colorMagenta, colorReset, errCount, float64(errCount)/float64(totalMutants)*100)

	score := (float64(killed) / float64(totalMutants)) * 100
	fmt.Printf("\n%sMutation Score: %.2f%%%s\n", colorBold, score, colorReset)

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

func runCandidate(ctx context.Context, c candidate, dir string, timeout time.Duration, fset *token.FileSet) (runner.Result, error) {
	originalBytes, err := os.ReadFile(c.filePath)
	if err != nil {
		return runner.Result{}, fmt.Errorf("reading source file: %w", err)
	}

	defer func() {
		// Restore the original file on disk to guarantee cleanliness
		_ = os.WriteFile(c.filePath, originalBytes, 0644)
	}()

	// Apply mutation to AST, formatting and writing it to the actual file path on disk
	err = astutil.ApplyAndWrite(fset, c.file, c.mutatedNode, c.filePath)
	if err != nil {
		return runner.Result{}, fmt.Errorf("applying mutation: %w", err)
	}

	res := runner.RunTests(ctx, dir, runner.Config{Timeout: timeout})

	res.Mutation = mutation.Mutation{
		Type:        mutation.MutationType(c.mutatorName),
		File:        c.filePath,
		Line:        c.line,
		Column:      c.column,
		Original:    c.original,
		Replacement: c.mutatedNode.Replacement,
		Status:      res.Status,
	}

	return res, nil
}
