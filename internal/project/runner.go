package project

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"syscall"

	"github.com/mokiat/gocrane/internal/logutil"
)

// NewRunner creates a new Runner with the specified arguments.
func NewRunner(args []string) *Runner {
	return &Runner{
		args: args,
	}
}

// Runner is responsible for running the built binary and stopping it when needed.
type Runner struct {
	args []string
}

// Run starts the program and returns a Process that can be used to stop it.
func (r *Runner) Run(ctx context.Context, binaryFile string) (*Process, error) {
	logger := log.New(log.Writer(), "[program]: ", log.Ltime|log.Lmsgprefix)

	runCtx, killFunc := context.WithCancel(ctx)
	cmd := exec.CommandContext(runCtx, binaryFile, r.args...)
	cmd.Stdout = logutil.ToWriter(logger)
	cmd.Stderr = logutil.ToWriter(logger)
	if err := cmd.Start(); err != nil {
		killFunc() // otherwise linter complains
		return nil, fmt.Errorf("failed to start program: %w", err)
	}
	return &Process{
		cmd:  cmd,
		kill: killFunc,
	}, nil
}

// Process represents a running program process, and can be used to stop it.
type Process struct {
	cmd  *exec.Cmd
	kill func()
}

// Stop attempts to stop the program gracefully, and if that fails it kills it.
func (p *Process) Stop(ctx context.Context) error {
	stopped := make(chan struct{})
	defer close(stopped)

	go func() {
		select {
		case <-ctx.Done():
			log.Println("Killing program, as it failed to shutdown gracefully...")
			p.kill()
		case <-stopped:
		}
	}()

	if err := p.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send sigterm signal to program: %w", err)
	}
	if err := p.cmd.Wait(); err != nil {
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			log.Printf("Program exited with non-zero exit code: %d", exitErr.ExitCode())
			return nil
		}
		return fmt.Errorf("failed to wait for program to stop: %w", err)
	}
	return nil
}
