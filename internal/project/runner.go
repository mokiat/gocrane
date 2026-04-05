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
func NewRunner(workDir string, args []string) *Runner {
	return &Runner{
		workDir: workDir,
		args:    args,
	}
}

// Runner is responsible for running the built binary and stopping it when needed.
type Runner struct {
	workDir string
	args    []string
}

// Run starts the program and returns a Process that can be used to stop it.
func (r *Runner) Run(binaryFile string) (*Process, error) {
	logger := log.New(log.Writer(), "[program]: ", log.Ltime|log.Lmsgprefix)

	ctxRun, killRun := context.WithCancel(context.Background())

	cmd := exec.CommandContext(ctxRun, binaryFile, r.args...)
	cmd.Dir = r.workDir
	cmd.Stdout = logutil.ToWriter(logger)
	cmd.Stderr = logutil.ToWriter(logger)
	if err := cmd.Start(); err != nil {
		killRun() // release resources associated with the context
		return nil, fmt.Errorf("failed to start program: %w", err)
	}

	return &Process{
		cmd:     cmd,
		ctxRun:  ctxRun,
		killRun: killRun,
	}, nil
}

// Process represents a running program process, and can be used to stop it.
type Process struct {
	cmd     *exec.Cmd
	ctxRun  context.Context
	killRun func()
}

// Stop attempts to stop the program gracefully, and if that fails it kills it.
//
// If the context is canceled before the program is stopped, the program will
// be killed forcefully.
func (p *Process) Stop(ctxShutdown context.Context) error {
	defer p.killRun() // release resources and indicate that process is stopped

	go func() {
		select {
		case <-ctxShutdown.Done():
			log.Println("Killing program, as it failed to shutdown gracefully...")
			p.killRun()
		case <-p.ctxRun.Done():
		}
	}()

	if err := p.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM signal to program: %w", err)
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
