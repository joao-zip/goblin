package config

import (
	"flag"
	"io"
	"strings"
	"time"
)

type Config struct {
	Dir       string
	Timeout   time.Duration
	Threshold float64
	Verbose   bool
	Mutators  []string
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

	fs := flag.NewFlagSet("gomutate", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	dirFlag := fs.String("dir", cfg.Dir, "Directory of the project to test")
	timeoutFlag := fs.Duration("timeout", cfg.Timeout, "Timeout for test execution")
	mutatorsFlag := fs.String("mutators", "", "Comma-separated list of mutators to run")
	thresholdFlag := fs.Float64("threshold", cfg.Threshold, "Minimum mutation score threshold")
	verboseFlag := fs.Bool("verbose", cfg.Verbose, "Enable verbose logging")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	cfg.Timeout = *timeoutFlag
	cfg.Threshold = *thresholdFlag
	cfg.Verbose = *verboseFlag

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
