package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joao-zip/goblin/internal/version"
)

type Config struct {
	Dir         string
	Timeout     time.Duration
	Threshold   float64
	Verbose     bool
	Mutators    []string
	Output      string
	HTML        string
	Workers     int
	ShowVersion bool
}

func Default() Config {
	return Config{
		Dir:       ".",
		Timeout:   10 * time.Second,
		Threshold: 0,
		Verbose:   false,
		Mutators:  nil,
	}
}

func FromArgs(args []string) (Config, error) {
	cfg := Default()

	fs := flag.NewFlagSet("goblin", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s", version.Banner())
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  goblin [flags] [directory]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fs.PrintDefaults()
	}

	dirFlag := fs.String("dir", cfg.Dir, "Directory of the project to test")
	timeoutFlag := fs.Duration("timeout", cfg.Timeout, "Timeout for test execution")
	mutatorsFlag := fs.String("mutators", "", "Comma-separated list of mutators to run")
	thresholdFlag := fs.Float64("threshold", cfg.Threshold, "Minimum mutation score threshold")
	verboseFlag := fs.Bool("verbose", cfg.Verbose, "Enable verbose logging")
	outputFlag := fs.String("output", "", "Output file path for JSON report")
	workersFlag := fs.Int("workers", 0, "Number of parallel workers (default: number of CPUs)")
	versionFlag := fs.Bool("version", false, "Print version information and exit")
	htmlFlag := fs.String("html", "", "Output file path for HTML report")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	if *versionFlag {
		cfg.ShowVersion = true
		return cfg, nil
	}

	cfg.Timeout = *timeoutFlag
	cfg.Threshold = *thresholdFlag
	cfg.Verbose = *verboseFlag
	cfg.Output = *outputFlag
	cfg.HTML = *htmlFlag
	cfg.Workers = *workersFlag

	if fs.NArg() > 0 {
		cfg.Dir = fs.Arg(0)
	} else {
		cfg.Dir = *dirFlag
	}

	if *mutatorsFlag != "" {
		parts := strings.Split(*mutatorsFlag, ",")
		cfg.Mutators = make([]string, len(parts))
		for i, part := range parts {
			cfg.Mutators[i] = strings.TrimSpace(part)
		}
	}

	return cfg, nil
}
