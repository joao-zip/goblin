package ast

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSource_ValidCode(t *testing.T) {
	src := `package main

func add(a, b int) int {
	return a + b
}
`
	file, fset, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource() error = %v", err)
	}
	if file == nil {
		t.Fatal("ParseSource() returned nil file")
	}
	if fset == nil {
		t.Fatal("ParseSource() returned nil fset")
	}
	if file.Name.Name != "main" {
		t.Errorf("package name = %q, want %q", file.Name.Name, "main")
	}
}

func TestParseSource_InvalidCode(t *testing.T) {
	src := `this is not valid go code`

	_, _, err := ParseSource(src)
	if err == nil {
		t.Fatal("ParseSource() should return error for invalid code")
	}
}

func TestParseSource_EmptyString(t *testing.T) {
	_, _, err := ParseSource("")
	if err == nil {
		t.Fatal("ParseSource() should return error for empty string")
	}
}

func TestParseFile_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "valid.go")
	content := []byte(`package example

func Hello() string { return "hello" }
`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	file, fset, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}
	if file == nil {
		t.Fatal("ParseFile() returned nil file")
	}
	if fset == nil {
		t.Fatal("ParseFile() returned nil fset")
	}
	if file.Name.Name != "example" {
		t.Errorf("package name = %q, want %q", file.Name.Name, "example")
	}
}

func TestParseFile_NonExistent(t *testing.T) {
	_, _, err := ParseFile("/nonexistent/path/file.go")
	if err == nil {
		t.Fatal("ParseFile() should return error for nonexistent file")
	}
}

func TestParseFile_InvalidContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.go")
	if err := os.WriteFile(path, []byte("not go code!!!"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, _, err := ParseFile(path)
	if err == nil {
		t.Fatal("ParseFile() should return error for invalid Go code")
	}
}

func TestParseDir_ValidDir(t *testing.T) {
	dir := t.TempDir()

	files := map[string]string{
		"a.go": `package example

func A() int { return 1 }
`,
		"b.go": `package example

func B() int { return 2 }
`,
	}

	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	parsed, fset, err := ParseDir(dir)
	if err != nil {
		t.Fatalf("ParseDir() error = %v", err)
	}
	if fset == nil {
		t.Fatal("ParseDir() returned nil fset")
	}
	if len(parsed) != 2 {
		t.Errorf("ParseDir() returned %d files, want 2", len(parsed))
	}
}

func TestParseDir_SkipsTestFiles(t *testing.T) {
	dir := t.TempDir()

	files := map[string]string{
		"code.go": `package example

func Code() {}
`,
		"code_test.go": `package example

func TestCode() {}
`,
	}

	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	parsed, _, err := ParseDir(dir)
	if err != nil {
		t.Fatalf("ParseDir() error = %v", err)
	}
	if len(parsed) != 1 {
		t.Errorf("ParseDir() returned %d files, want 1 (should skip _test.go)", len(parsed))
	}
}

func TestParseDir_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	parsed, _, err := ParseDir(dir)
	if err != nil {
		t.Fatalf("ParseDir() error = %v", err)
	}
	if len(parsed) != 0 {
		t.Errorf("ParseDir() returned %d files, want 0", len(parsed))
	}
}

func TestParseDir_NonExistent(t *testing.T) {
	_, _, err := ParseDir("/nonexistent/dir")
	if err == nil {
		t.Fatal("ParseDir() should return error for nonexistent dir")
	}
}
