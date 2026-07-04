package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/joao-zip/goblin/pkg/mutation"
)

// Config holds runner configuration.
type Config struct {
	Timeout time.Duration
}

// RunTests runs `go test ./...` in the given directory and classifies the result.
func RunTests(ctx context.Context, dir string, cfg Config) Result {
	start := time.Now()

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return Result{
			Status:   mutation.Error,
			Output:   fmt.Sprintf("directory does not exist: %s", dir),
			Duration: time.Since(start),
		}
	}

	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "test", "./...")
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)
	output := stdout.String() + stderr.String()

	status := classifyResult(ctx, err, output)

	return Result{
		Status:   status,
		Output:   output,
		Duration: duration,
	}
}

func classifyResult(ctx context.Context, err error, output string) mutation.MutationStatus {
	if err == nil {
		return mutation.Survived
	}

	if ctx.Err() == context.DeadlineExceeded {
		return mutation.Timeout
	}

	// go test returns exit code 1 on test failure, which is a normal "killed" result.
	// Build errors contain "build failed" or "[build failed]" in output.
	if isBuildError(output) {
		return mutation.Error
	}

	// Any other non-zero exit = tests failed = mutant killed
	return mutation.Killed
}

func isBuildError(output string) bool {
	return strings.Contains(output, "[build failed]") ||
		strings.Contains(output, "build failed") ||
		strings.Contains(output, "does not match any packages")
}
