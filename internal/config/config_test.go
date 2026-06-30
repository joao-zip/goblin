package config

import (
	"testing"
	"time"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Dir != "." {
		t.Errorf("Dir = %q, want %q", cfg.Dir, ".")
	}
	if cfg.Timeout != 10*time.Second {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, 10*time.Second)
	}
	if cfg.Threshold != 0 {
		t.Errorf("Threshold = %f, want 0", cfg.Threshold)
	}
	if cfg.Verbose {
		t.Error("Verbose = true, want false")
	}
	if cfg.Mutators != nil {
		t.Errorf("Mutators = %v, want nil", cfg.Mutators)
	}
	if cfg.Output != "" {
		t.Errorf("Output = %q, want empty", cfg.Output)
	}
	if cfg.Workers != 0 {
		t.Errorf("Workers = %d, want 0", cfg.Workers)
	}
}

func TestFromArgs_NoArgs(t *testing.T) {
	cfg, err := FromArgs([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := Default()
	if cfg.Dir != want.Dir {
		t.Errorf("Dir = %q, want %q", cfg.Dir, want.Dir)
	}
	if cfg.Timeout != want.Timeout {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, want.Timeout)
	}
	if cfg.Threshold != want.Threshold {
		t.Errorf("Threshold = %f, want %f", cfg.Threshold, want.Threshold)
	}
	if cfg.Verbose != want.Verbose {
		t.Errorf("Verbose = %v, want %v", cfg.Verbose, want.Verbose)
	}
	if cfg.Mutators != nil {
		t.Errorf("Mutators = %v, want nil", cfg.Mutators)
	}
}

func TestFromArgs_AllFlags(t *testing.T) {
	args := []string{
		"--dir", "/tmp/project",
		"--timeout", "30s",
		"--mutators", "arithmetic,comparison",
		"--threshold", "80",
		"--verbose",
	}
	cfg, err := FromArgs(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Dir != "/tmp/project" {
		t.Errorf("Dir = %q, want %q", cfg.Dir, "/tmp/project")
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, 30*time.Second)
	}
	if len(cfg.Mutators) != 2 || cfg.Mutators[0] != "arithmetic" || cfg.Mutators[1] != "comparison" {
		t.Errorf("Mutators = %v, want [arithmetic comparison]", cfg.Mutators)
	}
	if cfg.Threshold != 80 {
		t.Errorf("Threshold = %f, want 80", cfg.Threshold)
	}
	if !cfg.Verbose {
		t.Error("Verbose = false, want true")
	}
}

func TestFromArgs_PositionalArg(t *testing.T) {
	cfg, err := FromArgs([]string{"/tmp/myproject"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Dir != "/tmp/myproject" {
		t.Errorf("Dir = %q, want %q", cfg.Dir, "/tmp/myproject")
	}
}

func TestFromArgs_PositionalOverridesFlag(t *testing.T) {
	cfg, err := FromArgs([]string{"--dir", "/flag/path", "/positional/path"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Dir != "/positional/path" {
		t.Errorf("Dir = %q, want %q", cfg.Dir, "/positional/path")
	}
}

func TestFromArgs_InvalidFlag(t *testing.T) {
	_, err := FromArgs([]string{"--invalid"})
	if err == nil {
		t.Error("expected error for invalid flag")
	}
}

func TestFromArgs_EmptyMutators(t *testing.T) {
	cfg, err := FromArgs([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Mutators != nil {
		t.Errorf("Mutators = %v, want nil", cfg.Mutators)
	}
}

func TestFromArgs_OutputFlag(t *testing.T) {
	cfg, err := FromArgs([]string{"--output", "report.json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Output != "report.json" {
		t.Errorf("Output = %q, want %q", cfg.Output, "report.json")
	}
}

func TestFromArgs_SingleMutator(t *testing.T) {
	cfg, err := FromArgs([]string{"--mutators", "logical"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Mutators) != 1 || cfg.Mutators[0] != "logical" {
		t.Errorf("Mutators = %v, want [logical]", cfg.Mutators)
	}
}

func TestFromArgs_WorkersFlag(t *testing.T) {
	cfg, err := FromArgs([]string{"--workers", "4"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Workers != 4 {
		t.Errorf("Workers = %d, want 4", cfg.Workers)
	}
}

func TestFromArgs_WorkersDefault(t *testing.T) {
	cfg, err := FromArgs([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Workers != 0 {
		t.Errorf("Workers = %d, want 0", cfg.Workers)
	}
}
