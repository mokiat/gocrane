package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"

	"github.com/mokiat/gocrane/internal/command"
)

func main() {
	log.SetPrefix("[gocrane]: ")
	log.SetFlags(log.Ltime | log.Lmsgprefix)

	app := &cli.Command{
		Name:  "gocrane",
		Usage: "develop go applications in a docker environment",
		Flags: []cli.Flag{},
		Commands: []*cli.Command{
			command.Build(),
			command.Run(),
		},
	}

	appCtx, appStop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer appStop()
	if err := app.Run(appCtx, os.Args); err != nil {
		log.Fatalf("Crashed: %v", err)
	}
}
