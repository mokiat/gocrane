package pipeline

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mokiat/gocrane/internal/project"
)

func Run(
	ctx context.Context,
	runArgs []string,
	in Queue[BuildEvent],
	shutdownTimeout time.Duration,
) func() error {

	runner := project.NewRunner(runArgs)

	return func() error {
		var runningProcess *project.Process

		startProcess := func(path string) error {
			if runningProcess != nil {
				return fmt.Errorf("there is already a running process")
			}
			log.Printf("Starting new process...")
			process, err := runner.Run(context.Background(), path)
			if err != nil {
				return fmt.Errorf("failed to start process: %w", err)
			}
			log.Printf("Successfully started new process.")
			runningProcess = process
			return nil
		}

		stopProcess := func() error {
			if runningProcess == nil {
				return nil
			}
			log.Printf("Stopping running process (timeout: %s)...", shutdownTimeout)
			shutdownCtx, shutdownFunc := context.WithTimeout(context.Background(), shutdownTimeout)
			defer shutdownFunc()
			if err := runningProcess.Stop(shutdownCtx); err != nil {
				return fmt.Errorf("failed to stop process: %w", err)
			}
			log.Printf("Successfully stopped running process.")
			runningProcess = nil
			return nil
		}

		var buildEvent BuildEvent
		for in.Pop(ctx, &buildEvent) {
			if err := stopProcess(); err != nil {
				return err
			}
			if err := startProcess(buildEvent.Path); err != nil {
				return err
			}
		}
		return stopProcess()
	}
}
