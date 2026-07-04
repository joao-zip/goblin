package runner

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/joao-zip/goblin/pkg/mutation"
)

func TestRunAll_EmptyJobs(t *testing.T) {
	results := RunAll(context.Background(), nil, PoolConfig{
		Workers:    2,
		Timeout:    Config{Timeout: 5 * time.Second},
		ProjectDir: ".",
	})

	if results != nil {
		t.Errorf("expected nil for empty jobs, got %v", results)
	}
}

func TestRunAll_ReturnsAllResults(t *testing.T) {
	jobs := []Job{
		{ID: 1, RelPath: "a.go", OriginalData: []byte("a"), MutatedData: []byte("b")},
		{ID: 2, RelPath: "b.go", OriginalData: []byte("c"), MutatedData: []byte("d")},
		{ID: 3, RelPath: "c.go", OriginalData: []byte("e"), MutatedData: []byte("f")},
	}

	cfg := PoolConfig{
		Workers:    2,
		Timeout:    Config{Timeout: 5 * time.Second},
		ProjectDir: ".",
		TestFunc: func(_ context.Context, _ string, _ Config) Result {
			return Result{Status: mutation.Killed}
		},
	}

	results := RunAll(context.Background(), jobs, cfg)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	for _, r := range results {
		if r.Status != mutation.Killed {
			t.Errorf("expected Killed, got %s", r.Status)
		}
	}
}

func TestRunAll_ResultsOrderedByID(t *testing.T) {
	jobs := make([]Job, 10)
	for i := range jobs {
		jobs[i] = Job{
			ID:           i + 1,
			RelPath:      "test.go",
			OriginalData: []byte("orig"),
			MutatedData:  []byte("mutated"),
		}
	}

	cfg := PoolConfig{
		Workers:    4,
		Timeout:    Config{Timeout: 5 * time.Second},
		ProjectDir: ".",
		TestFunc: func(_ context.Context, _ string, _ Config) Result {
			return Result{Status: mutation.Killed}
		},
	}

	results := RunAll(context.Background(), jobs, cfg)

	if len(results) != 10 {
		t.Fatalf("expected 10 results, got %d", len(results))
	}

	for i, r := range results {
		if r.Mutation.ID != i+1 {
			t.Errorf("result[%d].ID = %d, want %d", i, r.Mutation.ID, i+1)
		}
	}
}

func TestRunAll_UsesAllWorkers(t *testing.T) {
	var maxConcurrent atomic.Int32
	var current atomic.Int32

	jobs := make([]Job, 8)
	for i := range jobs {
		jobs[i] = Job{
			ID:           i + 1,
			RelPath:      "test.go",
			OriginalData: []byte("orig"),
			MutatedData:  []byte("mutated"),
		}
	}

	cfg := PoolConfig{
		Workers:    4,
		Timeout:    Config{Timeout: 5 * time.Second},
		ProjectDir: ".",
		TestFunc: func(_ context.Context, _ string, _ Config) Result {
			cur := current.Add(1)
			for {
				old := maxConcurrent.Load()
				if cur <= old || maxConcurrent.CompareAndSwap(old, cur) {
					break
				}
			}
			time.Sleep(50 * time.Millisecond)
			current.Add(-1)
			return Result{Status: mutation.Killed}
		},
	}

	results := RunAll(context.Background(), jobs, cfg)

	if len(results) != 8 {
		t.Fatalf("expected 8 results, got %d", len(results))
	}

	if maxConcurrent.Load() < 2 {
		t.Errorf("expected at least 2 concurrent workers, got %d", maxConcurrent.Load())
	}
}

func TestRunAll_SingleWorker(t *testing.T) {
	jobs := []Job{
		{ID: 1, RelPath: "a.go", OriginalData: []byte("a"), MutatedData: []byte("b")},
		{ID: 2, RelPath: "b.go", OriginalData: []byte("c"), MutatedData: []byte("d")},
	}

	cfg := PoolConfig{
		Workers:    1,
		Timeout:    Config{Timeout: 5 * time.Second},
		ProjectDir: ".",
		TestFunc: func(_ context.Context, _ string, _ Config) Result {
			return Result{Status: mutation.Survived}
		},
	}

	results := RunAll(context.Background(), jobs, cfg)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	for _, r := range results {
		if r.Status != mutation.Survived {
			t.Errorf("expected Survived, got %s", r.Status)
		}
	}
}

func TestRunAll_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var started atomic.Int32

	jobs := make([]Job, 10)
	for i := range jobs {
		jobs[i] = Job{
			ID:           i + 1,
			RelPath:      "test.go",
			OriginalData: []byte("orig"),
			MutatedData:  []byte("mutated"),
		}
	}

	cfg := PoolConfig{
		Workers:    2,
		Timeout:    Config{Timeout: 5 * time.Second},
		ProjectDir: ".",
		TestFunc: func(ctx context.Context, _ string, _ Config) Result {
			n := started.Add(1)
			if n >= 3 {
				cancel()
			}
			time.Sleep(100 * time.Millisecond)
			if ctx.Err() != nil {
				return Result{Status: mutation.Error, Output: "context cancelled"}
			}
			return Result{Status: mutation.Killed}
		},
	}

	results := RunAll(ctx, jobs, cfg)

	if len(results) != 10 {
		t.Fatalf("expected 10 results, got %d", len(results))
	}
}

func TestRunAll_MutationMetadata(t *testing.T) {
	jobs := []Job{
		{
			ID:          1,
			RelPath:     "math.go",
			MutatorName: "arithmetic",
			Line:        10,
			Column:      5,
			Original:    "+",
			Replacement: "-",
			OriginalData: []byte("original"),
			MutatedData:  []byte("mutated"),
		},
	}

	cfg := PoolConfig{
		Workers:    1,
		Timeout:    Config{Timeout: 5 * time.Second},
		ProjectDir: ".",
		TestFunc: func(_ context.Context, _ string, _ Config) Result {
			return Result{Status: mutation.Killed}
		},
	}

	results := RunAll(context.Background(), jobs, cfg)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.Mutation.ID != 1 {
		t.Errorf("ID = %d, want 1", r.Mutation.ID)
	}
	if string(r.Mutation.Type) != "arithmetic" {
		t.Errorf("Type = %s, want arithmetic", r.Mutation.Type)
	}
	if r.Mutation.File != "math.go" {
		t.Errorf("File = %s, want math.go", r.Mutation.File)
	}
	if r.Mutation.Line != 10 {
		t.Errorf("Line = %d, want 10", r.Mutation.Line)
	}
	if r.Mutation.Original != "+" {
		t.Errorf("Original = %s, want +", r.Mutation.Original)
	}
	if r.Mutation.Replacement != "-" {
		t.Errorf("Replacement = %s, want -", r.Mutation.Replacement)
	}
}

func TestRunAll_WorkersCappedToJobs(t *testing.T) {
	jobs := []Job{
		{ID: 1, RelPath: "a.go", OriginalData: []byte("a"), MutatedData: []byte("b")},
	}

	cfg := PoolConfig{
		Workers:    100,
		Timeout:    Config{Timeout: 5 * time.Second},
		ProjectDir: ".",
		TestFunc: func(_ context.Context, _ string, _ Config) Result {
			return Result{Status: mutation.Killed}
		},
	}

	results := RunAll(context.Background(), jobs, cfg)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}
