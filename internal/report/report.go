package report

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/joao-zip/goblin/internal/runner"
	"github.com/joao-zip/goblin/pkg/mutation"
)

const (
	colorReset   = "\033[0m"
	colorBold    = "\033[1m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorMagenta = "\033[35m"
)

type Reporter interface {
	Report(w io.Writer, results []runner.Result) error
}

type TextReporter struct{}

type JSONReporter struct{}

type JSONReport struct {
	Mutants []JSONMutant `json:"mutants"`
	Summary JSONSummary  `json:"summary"`
}

type JSONMutant struct {
	ID          int    `json:"id"`
	Type        string `json:"type"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Column      int    `json:"column"`
	Original    string `json:"original"`
	Replacement string `json:"replacement"`
	Status      string `json:"status"`
	Duration    string `json:"duration"`
}

type JSONSummary struct {
	Total    int     `json:"total"`
	Killed   int     `json:"killed"`
	Survived int     `json:"survived"`
	Timeout  int     `json:"timeout"`
	Errors   int     `json:"errors"`
	Score    float64 `json:"score"`
}

func CalculateScore(results []runner.Result) float64 {
	if len(results) == 0 {
		return 0
	}

	var killed, survived int
	for _, r := range results {
		switch r.Status {
		case mutation.Killed:
			killed++
		case mutation.Survived:
			survived++
		}
	}

	denominator := killed + survived
	if denominator == 0 {
		return 0
	}

	return (float64(killed) / float64(denominator)) * 100
}

func (tr *TextReporter) Report(w io.Writer, results []runner.Result) error {
	if len(results) == 0 {
		_, err := fmt.Fprintln(w, "No mutations were generated.")
		return err
	}

	for _, r := range results {
		statusLabel := formatStatus(r.Status)
		fmt.Fprintf(w, "  %s %s:%d:%d  %s → %s  %s\n",
			statusLabel,
			r.Mutation.File,
			r.Mutation.Line,
			r.Mutation.Column,
			r.Mutation.Original,
			r.Mutation.Replacement,
			r.Duration,
		)
	}

	counts := tally(results)
	total := len(results)
	score := CalculateScore(results)

	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s--- Mutation Testing Results ---%s\n", colorBold, colorReset)
	fmt.Fprintf(w, "Total Mutants:  %d\n", total)
	fmt.Fprintf(w, "  %s✅ KILLED:%s      %d (%.1f%%)\n", colorGreen, colorReset, counts.killed, pct(counts.killed, total))
	fmt.Fprintf(w, "  %s❌ SURVIVED:%s    %d (%.1f%%)\n", colorRed, colorReset, counts.survived, pct(counts.survived, total))
	fmt.Fprintf(w, "  %s⏰ TIMEOUT:%s     %d (%.1f%%)\n", colorYellow, colorReset, counts.timeout, pct(counts.timeout, total))
	fmt.Fprintf(w, "  %s⚠️  ERROR:%s       %d (%.1f%%)\n", colorMagenta, colorReset, counts.errors, pct(counts.errors, total))
	fmt.Fprintf(w, "\n%sMutation Score: %.2f%%%s\n", colorBold, score, colorReset)

	return nil
}

func (jr *JSONReporter) Report(w io.Writer, results []runner.Result) error {
	counts := tally(results)

	report := JSONReport{
		Mutants: make([]JSONMutant, len(results)),
		Summary: JSONSummary{
			Total:    len(results),
			Killed:   counts.killed,
			Survived: counts.survived,
			Timeout:  counts.timeout,
			Errors:   counts.errors,
			Score:    CalculateScore(results),
		},
	}

	for i, r := range results {
		report.Mutants[i] = JSONMutant{
			ID:          r.Mutation.ID,
			Type:        string(r.Mutation.Type),
			File:        r.Mutation.File,
			Line:        r.Mutation.Line,
			Column:      r.Mutation.Column,
			Original:    r.Mutation.Original,
			Replacement: r.Mutation.Replacement,
			Status:      string(r.Status),
			Duration:    r.Duration.String(),
		}
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

type counts struct {
	killed   int
	survived int
	timeout  int
	errors   int
}

func tally(results []runner.Result) counts {
	var c counts
	for _, r := range results {
		switch r.Status {
		case mutation.Killed:
			c.killed++
		case mutation.Survived:
			c.survived++
		case mutation.Timeout:
			c.timeout++
		case mutation.Error:
			c.errors++
		}
	}
	return c
}

func pct(count, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(count) / float64(total) * 100
}

func formatStatus(status mutation.MutationStatus) string {
	switch status {
	case mutation.Killed:
		return colorGreen + "[KILLED]" + colorReset
	case mutation.Survived:
		return colorRed + "[SURVIVED]" + colorReset
	case mutation.Timeout:
		return colorYellow + "[TIMEOUT]" + colorReset
	case mutation.Error:
		return colorMagenta + "[ERROR]" + colorReset
	default:
		return "[UNKNOWN]"
	}
}
