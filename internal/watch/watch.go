package watch

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// Options configures the watch behavior
type Options struct {
	Interval time.Duration
	Clear    bool
}

// DefaultOptions returns the default watch options
func DefaultOptions() Options {
	return Options{
		Interval: 2 * time.Second,
		Clear:    true,
	}
}

// Watch executes a function repeatedly at the specified interval
// It clears the screen between executions if Clear is true
// Returns when the context is cancelled or the function returns an error
func Watch(ctx context.Context, opts Options, fn func() error) error {
	// Execute immediately
	if opts.Clear {
		clearScreen()
	}
	if err := fn(); err != nil {
		return err
	}

	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if opts.Clear {
				clearScreen()
			}
			fmt.Printf("\n[Last update: %s]\n\n", time.Now().Format("15:04:05"))
			if err := fn(); err != nil {
				return err
			}
		}
	}
}

// clearScreen clears the terminal screen
func clearScreen() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		// Silently ignore errors - clearing screen is not critical
		return
	}
}
