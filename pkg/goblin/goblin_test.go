package goblin_test

import (
	"testing"
	"time"

	"github.com/joao-zip/goblin/pkg/goblin"
)

func TestRun_CalculatorProject(t *testing.T) {
	result, err := goblin.Run(goblin.Options{
		Dir:     "../../testdata/calculator",
		Timeout: 30 * time.Second,
		Workers: 2,
	})
	if err != nil {
		t.Fatalf("goblin.Run() returned error: %v", err)
	}

	if result.Summary.Total == 0 {
		t.Error("expected total mutants to be > 0")
	}

	if result.Summary.Killed == 0 {
		t.Error("expected some killed mutants")
	}

	if result.Summary.Survived == 0 {
		t.Error("expected some survived mutants")
	}

	if result.Score <= 0 {
		t.Error("expected positive mutation score")
	}

	if result.Score > 100 {
		t.Error("mutation score should not exceed 100")
	}

	for _, m := range result.Mutants {
		if m.ID == 0 {
			t.Error("mutant ID should not be zero")
		}
		if m.File == "" {
			t.Error("mutant file should not be empty")
		}
		if m.Status == "" {
			t.Error("mutant status should not be empty")
		}
	}
}

func TestRun_Defaults(t *testing.T) {
	result, err := goblin.Run(goblin.Options{
		Dir: "../../testdata/calculator",
	})
	if err != nil {
		t.Fatalf("goblin.Run() with defaults returned error: %v", err)
	}

	if result.Summary.Total == 0 {
		t.Error("expected mutations with default options")
	}
}

func TestRun_ThresholdFailure(t *testing.T) {
	result, err := goblin.Run(goblin.Options{
		Dir:       "../../testdata/calculator",
		Threshold: 100.0,
		Workers:   2,
	})

	if err == nil {
		t.Fatal("expected ThresholdError when score < 100%")
	}

	thErr, ok := err.(*goblin.ThresholdError)
	if !ok {
		t.Fatalf("expected *ThresholdError, got %T: %v", err, err)
	}

	if thErr.Threshold != 100.0 {
		t.Errorf("expected threshold 100, got %f", thErr.Threshold)
	}

	if result == nil {
		t.Fatal("result should not be nil even on threshold failure")
	}
	if result.Summary.Total == 0 {
		t.Error("expected mutations even on threshold failure")
	}
}

func TestRun_FilterMutators(t *testing.T) {
	result, err := goblin.Run(goblin.Options{
		Dir:      "../../testdata/calculator",
		Mutators: []string{"arithmetic"},
		Workers:  2,
	})
	if err != nil {
		t.Fatalf("goblin.Run() with filtered mutators returned error: %v", err)
	}

	for _, m := range result.Mutants {
		if m.Type != "arithmetic" {
			t.Errorf("expected only arithmetic mutants, got type %q", m.Type)
		}
	}
}

func TestVersion(t *testing.T) {
	if goblin.Version == "" {
		t.Error("Version should not be empty")
	}
}
