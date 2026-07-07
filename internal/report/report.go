package report

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"time"

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

type HTMLReporter struct{}

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

// HTMLReportData holds the data passed to the HTML template.
type HTMLReportData struct {
	GeneratedAt string
	Summary     JSONSummary
	Files       []HTMLFile
}

// HTMLFile groups mutants by source file.
type HTMLFile struct {
	Name    string
	Mutants []HTMLMutant
}

// HTMLMutant holds per-mutant data for the HTML template.
type HTMLMutant struct {
	ID          int
	Line        int
	Column      int
	Type        string
	Original    string
	Replacement string
	Status      string
	Duration    string
}

// Report implements Reporter for HTMLReporter, writing a full interactive HTML page.
func (hr *HTMLReporter) Report(w io.Writer, results []runner.Result) error {
	counts := tally(results)
	score := CalculateScore(results)

	summary := JSONSummary{
		Total:    len(results),
		Killed:   counts.killed,
		Survived: counts.survived,
		Timeout:  counts.timeout,
		Errors:   counts.errors,
		Score:    score,
	}

	fileMap := make(map[string][]HTMLMutant)
	for _, r := range results {
		m := HTMLMutant{
			ID:          r.Mutation.ID,
			Line:        r.Mutation.Line,
			Column:      r.Mutation.Column,
			Type:        string(r.Mutation.Type),
			Original:    r.Mutation.Original,
			Replacement: r.Mutation.Replacement,
			Status:      string(r.Status),
			Duration:    r.Duration.Round(time.Millisecond).String(),
		}
		fileMap[r.Mutation.File] = append(fileMap[r.Mutation.File], m)
	}

	var files []HTMLFile
	for name, mutants := range fileMap {
		files = append(files, HTMLFile{Name: name, Mutants: mutants})
	}

	data := HTMLReportData{
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		Summary:     summary,
		Files:       files,
	}

	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"statusClass": func(s string) string {
			switch s {
			case "killed":
				return "killed"
			case "survived":
				return "survived"
			case "timeout":
				return "timeout"
			default:
				return "error"
			}
		},
		"scoreClass": func(score float64) string {
			switch {
			case score >= 80:
				return "score-good"
			case score >= 50:
				return "score-warn"
			default:
				return "score-bad"
			}
		},
	}).Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("parsing HTML template: %w", err)
	}

	return tmpl.Execute(w, data)
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Goblin — Mutation Report</title>
<style>
  @import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap');

  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

  :root {
    --bg:        #0d1117;
    --surface:   #161b22;
    --surface2:  #21262d;
    --border:    #30363d;
    --text:      #e6edf3;
    --muted:     #8b949e;
    --killed:    #3fb950;
    --killed-bg: #0d2b15;
    --survived:  #f85149;
    --survived-bg: #2d0f0e;
    --timeout:   #d29922;
    --timeout-bg: #2b1f07;
    --error:     #a371f7;
    --error-bg:  #1c1141;
    --accent:    #58a6ff;
    --radius:    10px;
  }

  body {
    font-family: 'Inter', system-ui, sans-serif;
    background: var(--bg);
    color: var(--text);
    min-height: 100vh;
    padding: 2rem 1rem;
  }

  .container { max-width: 1100px; margin: 0 auto; }

  /* Header */
  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 2rem;
    flex-wrap: wrap;
    gap: 1rem;
  }
  .logo { display: flex; align-items: center; gap: .75rem; }
  .logo h1 { font-size: 1.6rem; font-weight: 700; letter-spacing: -.5px; }
  .logo span { font-size: .85rem; color: var(--muted); }
  .generated { font-size: .8rem; color: var(--muted); }

  /* Summary cards */
  .summary {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
    gap: 1rem;
    margin-bottom: 2rem;
  }
  .card {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 1.25rem 1.5rem;
    display: flex;
    flex-direction: column;
    gap: .3rem;
  }
  .card-label { font-size: .75rem; text-transform: uppercase; letter-spacing: .08em; color: var(--muted); }
  .card-value { font-size: 2rem; font-weight: 700; line-height: 1; }
  .card-sub { font-size: .75rem; color: var(--muted); }

  .card.score-good .card-value { color: var(--killed); }
  .card.score-warn .card-value { color: var(--timeout); }
  .card.score-bad  .card-value { color: var(--survived); }
  .card.c-killed  .card-value  { color: var(--killed); }
  .card.c-survived .card-value { color: var(--survived); }
  .card.c-timeout .card-value  { color: var(--timeout); }
  .card.c-error   .card-value  { color: var(--error); }

  /* Filters */
  .filters {
    display: flex;
    gap: .5rem;
    margin-bottom: 1.5rem;
    flex-wrap: wrap;
    align-items: center;
  }
  .filters span { font-size: .85rem; color: var(--muted); margin-right: .25rem; }
  .filter-btn {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 20px;
    padding: .35rem .9rem;
    font-size: .8rem;
    font-family: inherit;
    color: var(--text);
    cursor: pointer;
    transition: all .15s;
  }
  .filter-btn:hover { border-color: var(--accent); color: var(--accent); }
  .filter-btn.active { background: var(--accent); border-color: var(--accent); color: #0d1117; font-weight: 600; }

  /* File sections */
  .file-section {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    margin-bottom: 1rem;
    overflow: hidden;
  }
  .file-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: .9rem 1.25rem;
    cursor: pointer;
    user-select: none;
    transition: background .15s;
  }
  .file-header:hover { background: var(--surface2); }
  .file-name {
    font-family: 'JetBrains Mono', monospace;
    font-size: .85rem;
    color: var(--accent);
    word-break: break-all;
  }
  .file-meta { display: flex; align-items: center; gap: .75rem; }
  .file-count { font-size: .8rem; color: var(--muted); }
  .chevron { color: var(--muted); font-size: .9rem; transition: transform .2s; }
  .file-section.open .chevron { transform: rotate(90deg); }

  .file-body { display: none; border-top: 1px solid var(--border); }
  .file-section.open .file-body { display: block; }

  /* Mutant rows */
  .mutant-row {
    display: grid;
    grid-template-columns: 2.5rem 5rem 1fr 1fr auto auto;
    align-items: center;
    gap: .75rem;
    padding: .7rem 1.25rem;
    border-bottom: 1px solid var(--border);
    transition: background .1s;
    font-size: .85rem;
  }
  .mutant-row:last-child { border-bottom: none; }
  .mutant-row:hover { background: var(--surface2); }

  .mutant-id { color: var(--muted); font-family: 'JetBrains Mono', monospace; font-size: .75rem; }
  .mutant-loc { font-family: 'JetBrains Mono', monospace; font-size: .75rem; color: var(--muted); }
  .mutant-type { font-size: .72rem; background: var(--surface2); border: 1px solid var(--border); border-radius: 4px; padding: .1rem .4rem; color: var(--muted); white-space: nowrap; }

  .diff {
    display: flex;
    align-items: center;
    gap: .4rem;
    font-family: 'JetBrains Mono', monospace;
    font-size: .82rem;
    flex-wrap: wrap;
  }
  .diff-orig  { background: var(--survived-bg); color: var(--survived); padding: .1rem .35rem; border-radius: 4px; }
  .diff-arrow { color: var(--muted); }
  .diff-new   { background: var(--killed-bg); color: var(--killed); padding: .1rem .35rem; border-radius: 4px; }

  .badge {
    padding: .2rem .55rem;
    border-radius: 20px;
    font-size: .72rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: .05em;
    white-space: nowrap;
  }
  .badge.killed   { background: var(--killed-bg);   color: var(--killed); }
  .badge.survived { background: var(--survived-bg); color: var(--survived); }
  .badge.timeout  { background: var(--timeout-bg);  color: var(--timeout); }
  .badge.error    { background: var(--error-bg);    color: var(--error); }

  .mutant-dur { font-size: .72rem; color: var(--muted); white-space: nowrap; }

  /* Empty state */
  .empty { text-align: center; padding: 3rem; color: var(--muted); font-size: .95rem; }

  /* Responsive */
  @media (max-width: 640px) {
    .mutant-row { grid-template-columns: 1fr 1fr; row-gap: .4rem; }
    .mutant-id, .mutant-dur { display: none; }
  }
</style>
</head>
<body>
<div class="container">
  <header>
    <div class="logo">
      <h1>Goblin</h1>
      <span>Mutation Report</span>
    </div>
    <span class="generated">Generated at {{.GeneratedAt}}</span>
  </header>

  <div class="summary">
    <div class="card {{scoreClass .Summary.Score}}">
      <span class="card-label">Mutation Score</span>
      <span class="card-value">{{printf "%.1f" .Summary.Score}}%</span>
      <span class="card-sub">killed / (killed + survived)</span>
    </div>
    <div class="card">
      <span class="card-label">Total Mutants</span>
      <span class="card-value" style="color:var(--accent)">{{.Summary.Total}}</span>
    </div>
    <div class="card c-killed">
      <span class="card-label">Killed</span>
      <span class="card-value">{{.Summary.Killed}}</span>
    </div>
    <div class="card c-survived">
      <span class="card-label">Survived</span>
      <span class="card-value">{{.Summary.Survived}}</span>
    </div>
    <div class="card c-timeout">
      <span class="card-label">Timeout</span>
      <span class="card-value">{{.Summary.Timeout}}</span>
    </div>
    <div class="card c-error">
      <span class="card-label">Errors</span>
      <span class="card-value">{{.Summary.Errors}}</span>
    </div>
  </div>

  <div class="filters">
    <span>Filter:</span>
    <button class="filter-btn active" data-filter="all">All</button>
    <button class="filter-btn" data-filter="killed">Killed</button>
    <button class="filter-btn" data-filter="survived">Survived</button>
    <button class="filter-btn" data-filter="timeout">Timeout</button>
    <button class="filter-btn" data-filter="error">Error</button>
  </div>

  {{if .Files}}
  {{range .Files}}
  <div class="file-section open" data-file>
    <div class="file-header" onclick="toggleFile(this)">
      <span class="file-name">{{.Name}}</span>
      <div class="file-meta">
        <span class="file-count">{{len .Mutants}} mutant(s)</span>
        <span class="chevron">▶</span>
      </div>
    </div>
    <div class="file-body">
      {{range .Mutants}}
      <div class="mutant-row" data-status="{{.Status}}">
        <span class="mutant-id">#{{.ID}}</span>
        <span class="mutant-loc">{{.Line}}:{{.Column}}</span>
        <span class="mutant-type">{{.Type}}</span>
        <div class="diff">
          <span class="diff-orig">{{.Original}}</span>
          <span class="diff-arrow">→</span>
          <span class="diff-new">{{.Replacement}}</span>
        </div>
        <span class="badge {{statusClass .Status}}">{{.Status}}</span>
        <span class="mutant-dur">{{.Duration}}</span>
      </div>
      {{end}}
    </div>
  </div>
  {{end}}
  {{else}}
  <div class="empty">No mutations were generated.</div>
  {{end}}
</div>

<script>
  function toggleFile(header) {
    header.closest('.file-section').classList.toggle('open');
  }

  document.querySelectorAll('.filter-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      const filter = btn.dataset.filter;

      document.querySelectorAll('[data-file]').forEach(section => {
        const rows = section.querySelectorAll('.mutant-row');
        let visible = 0;
        rows.forEach(row => {
          const match = filter === 'all' || row.dataset.status === filter;
          row.style.display = match ? '' : 'none';
          if (match) visible++;
        });
        section.style.display = visible === 0 ? 'none' : '';
        section.querySelector('.file-count').textContent = visible + ' mutant(s)';
      });
    });
  });
</script>
</body>
</html>`
