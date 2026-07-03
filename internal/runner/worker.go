package runner

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"

	"github.com/joao-zip/gomutate/pkg/mutation"
)

type Job struct {
	ID           int
	RelPath      string
	OriginalData []byte
	MutatedData  []byte
	MutatorName  string
	Line         int
	Column       int
	Original     string
	Replacement  string
}

type PoolConfig struct {
	Workers    int
	Timeout    Config
	ProjectDir string
	TestFunc   func(ctx context.Context, dir string, cfg Config) Result
}

func RunAll(ctx context.Context, jobs []Job, cfg PoolConfig) []Result {
	if len(jobs) == 0 {
		return nil
	}

	workers := cfg.Workers
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	if workers > len(jobs) {
		workers = len(jobs)
	}

	testFn := cfg.TestFunc
	if testFn == nil {
		testFn = RunTests
	}

	jobsCh := make(chan Job, len(jobs))
	resultsCh := make(chan Result, len(jobs))

	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker(ctx, jobsCh, resultsCh, cfg, testFn)
		}()
	}

	for _, j := range jobs {
		jobsCh <- j
	}
	close(jobsCh)

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	results := make([]Result, 0, len(jobs))
	for r := range resultsCh {
		results = append(results, r)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Mutation.ID < results[j].Mutation.ID
	})

	return results
}

func worker(ctx context.Context, jobs <-chan Job, results chan<- Result, cfg PoolConfig, testFn func(context.Context, string, Config) Result) {
	tmpDir, err := copyDir(cfg.ProjectDir)
	if err != nil {
		for j := range jobs {
			results <- Result{
				Mutation: mutation.Mutation{ID: j.ID, Status: mutation.Error},
				Status:   mutation.Error,
				Output:   fmt.Sprintf("failed to create temp project copy: %v", err),
			}
		}
		return
	}
	defer os.RemoveAll(tmpDir)

	for j := range jobs {
		select {
		case <-ctx.Done():
			results <- Result{
				Mutation: buildMutation(j, mutation.Error),
				Status:   mutation.Error,
				Output:   "context cancelled",
			}
			continue
		default:
		}

		filePath := filepath.Join(tmpDir, j.RelPath)

		if err := os.WriteFile(filePath, j.MutatedData, 0644); err != nil {
			results <- Result{
				Mutation: buildMutation(j, mutation.Error),
				Status:   mutation.Error,
				Output:   fmt.Sprintf("writing mutated file: %v", err),
			}
			continue
		}

		res := testFn(ctx, tmpDir, cfg.Timeout)

		_ = os.WriteFile(filePath, j.OriginalData, 0644)

		res.Mutation = buildMutation(j, res.Status)
		results <- res
	}
}

func buildMutation(j Job, status mutation.MutationStatus) mutation.Mutation {
	return mutation.Mutation{
		ID:          j.ID,
		Type:        mutation.MutationType(j.MutatorName),
		File:        j.RelPath,
		Line:        j.Line,
		Column:      j.Column,
		Original:    j.Original,
		Replacement: j.Replacement,
		Status:      status,
	}
}

func copyDir(src string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "gomutate-worker-*")
	if err != nil {
		return "", fmt.Errorf("creating temp dir: %w", err)
	}

	err = filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dst := filepath.Join(tmpDir, rel)

		if d.IsDir() {
			return os.MkdirAll(dst, 0755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(dst, data, 0644)
	})

	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("copying project dir: %w", err)
	}

	return tmpDir, nil
}
