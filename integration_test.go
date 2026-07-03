package main_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestE2E_Calculator(t *testing.T) {
	// 1. Build the gomutate binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "gomutate")

	cmdBuild := exec.Command("go", "build", "-o", binaryPath, "./cmd/gomutate")
	cmdBuild.Env = append(os.Environ(), "GOTOOLCHAIN=auto")
	if out, err := cmdBuild.CombinedOutput(); err != nil {
		t.Fatalf("failed to build gomutate binary: %v\nOutput: %s", err, string(out))
	}

	// 2. Run integration test without threshold, with output json
	reportPath := filepath.Join(tmpDir, "report.json")
	cmdRun := exec.Command(binaryPath, "--dir", "./testdata/calculator", "--output", reportPath)
	cmdRun.Env = append(os.Environ(), "GOTOOLCHAIN=auto")

	var stdout, stderr bytes.Buffer
	cmdRun.Stdout = &stdout
	cmdRun.Stderr = &stderr

	err := cmdRun.Run()
	if err != nil {
		t.Fatalf("gomutate failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify stdout contains expected content
	outputStr := stdout.String()
	if !bytes.Contains(stdout.Bytes(), []byte("Mutation Testing Results")) {
		t.Errorf("stdout does not contain summary, got: %s", outputStr)
	}

	// Verify JSON file exists and contains survived and killed mutants
	reportData, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("failed to read JSON report: %v", err)
	}

	type Report struct {
		Mutants []struct {
			ID          int    `json:"id"`
			Type        string `json:"type"`
			File        string `json:"file"`
			Line        int    `json:"line"`
			Column      int    `json:"column"`
			Original    string `json:"original"`
			Replacement string `json:"replacement"`
			Status      string `json:"status"`
		} `json:"mutants"`
		Summary struct {
			Total    int     `json:"total"`
			Killed   int     `json:"killed"`
			Survived int     `json:"survived"`
			Timeout  int     `json:"timeout"`
			Errors   int     `json:"errors"`
			Score    float64 `json:"score"`
		} `json:"summary"`
	}

	var rep Report
	if err := json.Unmarshal(reportData, &rep); err != nil {
		t.Fatalf("failed to parse JSON report: %v", err)
	}

	if rep.Summary.Total == 0 {
		t.Error("expected total mutants to be > 0")
	}

	if rep.Summary.Killed == 0 {
		t.Error("expected some killed mutants (Add, Subtract)")
	}
	if rep.Summary.Survived == 0 {
		t.Error("expected some survived mutants (Multiply, Divide)")
	}

	// 3. Test threshold failure
	cmdThreshold := exec.Command(binaryPath, "--dir", "./testdata/calculator", "--threshold", "100")
	cmdThreshold.Env = append(os.Environ(), "GOTOOLCHAIN=auto")

	var stdoutT, stderrT bytes.Buffer
	cmdThreshold.Stdout = &stdoutT
	cmdThreshold.Stderr = &stderrT

	errT := cmdThreshold.Run()
	if errT == nil {
		t.Fatal("expected failure exit code when score is below threshold, but got 0")
	}

	if !bytes.Contains(stdoutT.Bytes(), []byte("Failure: Mutation score is below the required threshold")) {
		t.Errorf("expected threshold failure message in stdout, got: %s", stdoutT.String())
	}
}
