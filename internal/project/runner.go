package project

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/mokiat/gocrane/internal/logutil"
)

func NewRunner(args []string) *Runner {
	return &Runner{
		args: args,
	}
}

type Runner struct {
	args []string
}

func (r *Runner) Run(ctx context.Context, path string) (*Process, error) {
	logger := log.New(log.Writer(), "[program]: ", log.Ltime|log.Lmsgprefix)

	runCtx, killFunc := context.WithCancel(ctx)
	cmd := exec.CommandContext(runCtx, path, r.args...)
	cmd.Stdout = logutil.ToWriter(logger)
	cmd.Stderr = logutil.ToWriter(logger)
	if err := cmd.Start(); err != nil {
		killFunc() // otherwise linter complains
		return nil, fmt.Errorf("failed to start program: %w", err)
	}
	return &Process{
		process: cmd.Process,
		kill:    killFunc,
	}, nil
}

type Process struct {
	process *os.Process
	kill    func()
}

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

	if err := p.process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send sigterm signal to program: %w", err)
	}
	state, err := p.process.Wait()
	if err != nil {
		return fmt.Errorf("failed to wait for program to stop: %w", err)
	}
	if !state.Success() {
		log.Printf("Program exited with non-zero exit code: %d", state.ExitCode())
	}
	return nil
}
