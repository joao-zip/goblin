package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joao-zip/gomutate/pkg/mutation"
)

func TestRunTests_Passing(t *testing.T) {
	dir := createTestProject(t, `package example

func Add(a, b int) int { return a + b }
`, `package example

import "testing"

func TestAdd(t *testing.T) {
	if Add(1, 2) != 3 {
		t.Error("expected 3")
	}
}
`)

	ctx := context.Background()
	result := RunTests(ctx, dir, Config{Timeout: 30 * time.Second})

	if result.Status != mutation.Survived {
		t.Errorf("Status = %s, want %s (tests pass = mutant survived)", result.Status, mutation.Survived)
	}
	if result.Duration <= 0 {
		t.Error("Duration should be > 0")
	}
}

func TestRunTests_Failing(t *testing.T) {
	dir := createTestProject(t, `package example

func Add(a, b int) int { return a - b }
`, `package example

import "testing"

func TestAdd(t *testing.T) {
	if Add(1, 2) != 3 {
		t.Error("expected 3")
	}
}
`)

	ctx := context.Background()
	result := RunTests(ctx, dir, Config{Timeout: 30 * time.Second})

	if result.Status != mutation.Killed {
		t.Errorf("Status = %s, want %s (tests fail = mutant killed)", result.Status, mutation.Killed)
	}
}

func TestRunTests_Timeout(t *testing.T) {
	dir := createTestProject(t, `package example

func Slow() {}
`, `package example

import (
	"testing"
	"time"
)

func TestSlow(t *testing.T) {
	time.Sleep(10 * time.Second)
}
`)

	ctx := context.Background()
	result := RunTests(ctx, dir, Config{Timeout: 500 * time.Millisecond})

	if result.Status != mutation.Timeout {
		t.Errorf("Status = %s, want %s", result.Status, mutation.Timeout)
	}
}

func TestRunTests_BuildError(t *testing.T) {
	dir := createTestProject(t, `package example

func Bad( { invalid go
`, `package example

import "testing"

func TestBad(t *testing.T) {}
`)

	ctx := context.Background()
	result := RunTests(ctx, dir, Config{Timeout: 30 * time.Second})

	if result.Status != mutation.Error {
		t.Errorf("Status = %s, want %s", result.Status, mutation.Error)
	}
}

func TestRunTests_NonExistentDir(t *testing.T) {
	ctx := context.Background()
	result := RunTests(ctx, "/nonexistent/dir", Config{Timeout: 30 * time.Second})

	if result.Status != mutation.Error {
		t.Errorf("Status = %s, want %s", result.Status, mutation.Error)
	}
}

func TestRunTests_OutputCaptured(t *testing.T) {
	dir := createTestProject(t, `package example

func Add(a, b int) int { return a + b }
`, `package example

import "testing"

func TestAdd(t *testing.T) {
	if Add(1, 2) != 3 {
		t.Error("expected 3")
	}
}
`)

	ctx := context.Background()
	result := RunTests(ctx, dir, Config{Timeout: 30 * time.Second})

	if result.Output == "" {
		t.Error("Output should not be empty")
	}
}

// createTestProject creates a minimal Go project in a temp directory.
func createTestProject(t *testing.T, source, testCode string) string {
	t.Helper()
	dir := t.TempDir()

	goMod := "module testproject\n\ngo 1.21\n"
	writeTestFile(t, dir, "go.mod", goMod)
	writeTestFile(t, dir, "code.go", source)
	writeTestFile(t, dir, "code_test.go", testCode)

	return dir
}

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing %s: %v", name, err)
	}
}
