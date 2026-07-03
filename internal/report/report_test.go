package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/joao-zip/gomutate/internal/runner"
	"github.com/joao-zip/gomutate/pkg/mutation"
)

func sampleResults() []runner.Result {
	return []runner.Result{
		{
			Mutation: mutation.Mutation{
				ID:          1,
				Type:        mutation.Arithmetic,
				File:        "math.go",
				Line:        10,
				Column:      5,
				Original:    "+",
				Replacement: "-",
				Status:      mutation.Killed,
			},
			Status:   mutation.Killed,
			Duration: 200 * time.Millisecond,
		},
		{
			Mutation: mutation.Mutation{
				ID:          2,
				Type:        mutation.Comparison,
				File:        "cmp.go",
				Line:        20,
				Column:      8,
				Original:    "==",
				Replacement: "!=",
				Status:      mutation.Survived,
			},
			Status:   mutation.Survived,
			Duration: 150 * time.Millisecond,
		},
		{
			Mutation: mutation.Mutation{
				ID:          3,
				Type:        mutation.Logical,
				File:        "logic.go",
				Line:        30,
				Column:      12,
				Original:    "&&",
				Replacement: "||",
				Status:      mutation.Timeout,
			},
			Status:   mutation.Timeout,
			Duration: 10 * time.Second,
		},
		{
			Mutation: mutation.Mutation{
				ID:          4,
				Type:        mutation.Arithmetic,
				File:        "math.go",
				Line:        15,
				Column:      5,
				Original:    "*",
				Replacement: "/",
				Status:      mutation.Error,
			},
			Status:   mutation.Error,
			Duration: 50 * time.Millisecond,
		},
	}
}

func TestTextReporter_ImplementsReporter(t *testing.T) {
	var _ Reporter = &TextReporter{}
}

func TestJSONReporter_ImplementsReporter(t *testing.T) {
	var _ Reporter = &JSONReporter{}
}

func TestTextReporter_Report(t *testing.T) {
	var buf bytes.Buffer
	reporter := &TextReporter{}
	results := sampleResults()

	err := reporter.Report(&buf, results)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	expectedPhrases := []string{
		"KILLED",
		"SURVIVED",
		"TIMEOUT",
		"ERROR",
		"Mutation Score",
		"math.go",
		"cmp.go",
		"logic.go",
	}
	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("output missing %q", phrase)
		}
	}
}

func TestTextReporter_EmptyResults(t *testing.T) {
	var buf bytes.Buffer
	reporter := &TextReporter{}

	err := reporter.Report(&buf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No mutations") {
		t.Error("expected empty results message")
	}
}

func TestTextReporter_AllKilled(t *testing.T) {
	var buf bytes.Buffer
	reporter := &TextReporter{}
	results := []runner.Result{
		{
			Mutation: mutation.Mutation{
				ID: 1, Original: "+", Replacement: "-", Status: mutation.Killed,
			},
			Status: mutation.Killed,
		},
		{
			Mutation: mutation.Mutation{
				ID: 2, Original: "*", Replacement: "/", Status: mutation.Killed,
			},
			Status: mutation.Killed,
		},
	}

	err := reporter.Report(&buf, results)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "100.00%") {
		t.Errorf("expected 100%% score, got: %s", output)
	}
}

func TestJSONReporter_Report(t *testing.T) {
	var buf bytes.Buffer
	reporter := &JSONReporter{}
	results := sampleResults()

	err := reporter.Report(&buf, results)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed JSONReport
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(parsed.Mutants) != 4 {
		t.Errorf("Mutants count = %d, want 4", len(parsed.Mutants))
	}

	if parsed.Summary.Total != 4 {
		t.Errorf("Total = %d, want 4", parsed.Summary.Total)
	}
	if parsed.Summary.Killed != 1 {
		t.Errorf("Killed = %d, want 1", parsed.Summary.Killed)
	}
	if parsed.Summary.Survived != 1 {
		t.Errorf("Survived = %d, want 1", parsed.Summary.Survived)
	}
	if parsed.Summary.Timeout != 1 {
		t.Errorf("Timeout = %d, want 1", parsed.Summary.Timeout)
	}
	if parsed.Summary.Errors != 1 {
		t.Errorf("Errors = %d, want 1", parsed.Summary.Errors)
	}
	if parsed.Summary.Score != 50.0 {
		t.Errorf("Score = %f, want 50.0", parsed.Summary.Score)
	}
}

func TestJSONReporter_EmptyResults(t *testing.T) {
	var buf bytes.Buffer
	reporter := &JSONReporter{}

	err := reporter.Report(&buf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed JSONReport
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if parsed.Summary.Total != 0 {
		t.Errorf("Total = %d, want 0", parsed.Summary.Total)
	}
	if parsed.Summary.Score != 0 {
		t.Errorf("Score = %f, want 0", parsed.Summary.Score)
	}
}

func TestCalculateScore_KilledAndSurvived(t *testing.T) {
	results := sampleResults()
	score := CalculateScore(results)

	if score != 50.0 {
		t.Errorf("Score = %f, want 50.0", score)
	}
}

func TestCalculateScore_NoResults(t *testing.T) {
	score := CalculateScore(nil)
	if score != 0 {
		t.Errorf("Score = %f, want 0", score)
	}
}

func TestCalculateScore_OnlyKilled(t *testing.T) {
	results := []runner.Result{
		{Status: mutation.Killed},
		{Status: mutation.Killed},
	}

	score := CalculateScore(results)
	if score != 100.0 {
		t.Errorf("Score = %f, want 100.0", score)
	}
}

func TestCalculateScore_OnlySurvived(t *testing.T) {
	results := []runner.Result{
		{Status: mutation.Survived},
		{Status: mutation.Survived},
	}

	score := CalculateScore(results)
	if score != 0 {
		t.Errorf("Score = %f, want 0", score)
	}
}
