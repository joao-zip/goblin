package ast

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joao-zip/gomutate/internal/mutator"
)

func TestWriteFile(t *testing.T) {
	src := `package example

func Add(a, b int) int {
	return a + b
}
`
	file, fset, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource() error = %v", err)
	}

	outPath := filepath.Join(t.TempDir(), "out.go")
	if err := WriteFile(fset, file, outPath); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}

	if !strings.Contains(string(content), "func Add") {
		t.Error("output should contain func Add")
	}
	if !strings.Contains(string(content), "a + b") {
		t.Error("output should contain a + b")
	}
}

func TestWriteFile_OutputIsValidGo(t *testing.T) {
	src := `package example

func Calc(x, y int) int {
	if x > 0 && y < 10 {
		return x + y
	}
	return x - y
}
`
	file, fset, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource() error = %v", err)
	}

	outPath := filepath.Join(t.TempDir(), "out.go")
	if err := WriteFile(fset, file, outPath); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Re-parse to verify it's valid Go
	_, _, err = ParseFile(outPath)
	if err != nil {
		t.Errorf("output is not valid Go: %v", err)
	}
}

func TestWriteFile_InvalidPath(t *testing.T) {
	src := `package example`
	file, fset, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource() error = %v", err)
	}

	err = WriteFile(fset, file, "/nonexistent/dir/file.go")
	if err == nil {
		t.Error("WriteFile() should return error for invalid path")
	}
}

func TestApplyAndWrite(t *testing.T) {
	src := `package example

func Add(a, b int) int {
	return a + b
}
`
	file, fset, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource() error = %v", err)
	}

	nodes := FindMutableNodes(file, fset)
	if len(nodes) == 0 {
		t.Fatal("expected mutable nodes")
	}

	m := &mutator.ArithmeticMutator{}
	mutations := m.Mutate(nodes[0].Node)
	if len(mutations) == 0 {
		t.Fatal("expected mutations")
	}

	mut := mutations[0] // + -> -
	outPath := filepath.Join(t.TempDir(), "mutated.go")

	if err := ApplyAndWrite(fset, file, mut, outPath); err != nil {
		t.Fatalf("ApplyAndWrite() error = %v", err)
	}

	// Verify the written file has the mutation
	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	if !strings.Contains(string(content), "a - b") {
		t.Errorf("output should contain mutated expression 'a - b', got:\n%s", content)
	}

	// Verify the original AST was rolled back
	outOriginal := filepath.Join(t.TempDir(), "original.go")
	if err := WriteFile(fset, file, outOriginal); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	original, _ := os.ReadFile(outOriginal)
	if !strings.Contains(string(original), "a + b") {
		t.Errorf("AST should be rolled back to 'a + b', got:\n%s", original)
	}
}

func TestApplyAndWrite_OutputIsValidGo(t *testing.T) {
	src := `package example

func Div(a, b int) int {
	return a / b
}
`
	file, fset, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource() error = %v", err)
	}

	nodes := FindMutableNodes(file, fset)
	m := &mutator.ArithmeticMutator{}
	mutations := m.Mutate(nodes[0].Node)

	outPath := filepath.Join(t.TempDir(), "mutated.go")
	if err := ApplyAndWrite(fset, file, mutations[0], outPath); err != nil {
		t.Fatalf("ApplyAndWrite() error = %v", err)
	}

	_, _, err = ParseFile(outPath)
	if err != nil {
		t.Errorf("mutated output is not valid Go: %v", err)
	}
}
