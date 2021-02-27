package command

import (
	"context"
	"log"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/mokiat/gocrane/internal/flag"
)

func Run() *cli.Command {
	var cfg runConfig
	return &cli.Command{
		Name: "run",
		Flags: []cli.Flag{
			newVerboseFlag(&cfg.Verbose),
			newSourcesFlag(&cfg.Sources),
			newResourcesFlag(&cfg.Resources),
			newExcludesFlag(&cfg.Excludes),
			newMainFlag(&cfg.MainDir),
			newBinaryFlag(&cfg.BinaryFile, false),
			newDigestFlag(&cfg.DigestFile),
			newBuildArgs(&cfg.BuildArgs),
			newRunArgs(&cfg.RunArgs),
			newShutdownTimeoutFlag(&cfg.ShutdownTimeout),
		},
		Action: func(c *cli.Context) error {
			return run(c.Context, cfg)
		},
	}
}

type runConfig struct {
	Verbose         bool
	Sources         cli.StringSlice
	Resources       cli.StringSlice
	Excludes        cli.StringSlice
	MainDir         string
	BinaryFile      string
	DigestFile      string
	BuildArgs       flag.ShlexStringSlice
	RunArgs         flag.ShlexStringSlice
	ShutdownTimeout time.Duration
}

func run(ctx context.Context, cfg runConfig) error {
	log.Printf("verbose: %v", cfg.Verbose)
	log.Printf("sources: %v", cfg.Sources.Value())
	log.Printf("resources: %v", cfg.Resources.Value())
	log.Printf("excludes: %v", cfg.Excludes.Value())
	log.Printf("main dir: %v", cfg.MainDir)
	log.Printf("binary file: %v", cfg.BinaryFile)
	log.Printf("digest file: %v", cfg.DigestFile)
	log.Printf("build args: %v", cfg.BuildArgs)
	log.Printf("run args: %v", cfg.RunArgs)
	log.Printf("shutdown timeout: %v", cfg.ShutdownTimeout)
	return nil
}
