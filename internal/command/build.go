package command

import (
	"context"
	"log"

	"github.com/mokiat/gocrane/internal/flag"
	"github.com/urfave/cli/v2"
)

func Build() *cli.Command {
	var cfg buildConfig
	return &cli.Command{
		Name: "build",
		Flags: []cli.Flag{
			newVerboseFlag(&cfg.Verbose),
			newSourcesFlag(&cfg.Sources),
			newResourcesFlag(&cfg.Resources),
			newExcludesFlag(&cfg.Excludes),
			newMainFlag(&cfg.MainDir),
			newBinaryFlag(&cfg.BinaryFile, true),
			newDigestFlag(&cfg.DigestFile),
			newBuildArgs(&cfg.BuildArgs),
		},
		Action: func(c *cli.Context) error {
			return build(c.Context, cfg)
		},
	}
}

type buildConfig struct {
	Verbose    bool
	Sources    cli.StringSlice
	Resources  cli.StringSlice
	Excludes   cli.StringSlice
	MainDir    string
	BinaryFile string
	DigestFile string
	BuildArgs  flag.ShlexStringSlice
}

func build(ctx context.Context, cfg buildConfig) error {
	log.Printf("verbose: %v", cfg.Verbose)
	log.Printf("sources: %v", cfg.Sources.Value())
	log.Printf("resources: %v", cfg.Resources.Value())
	log.Printf("excludes: %v", cfg.Excludes.Value())
	log.Printf("main dir: %v", cfg.MainDir)
	log.Printf("binary file: %v", cfg.BinaryFile)
	log.Printf("digest file: %v", cfg.DigestFile)
	log.Printf("build args: %v", cfg.BuildArgs)
	return nil
}
