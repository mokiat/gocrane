package command

import (
	"context"
	"fmt"
	"log"

	"github.com/urfave/cli/v2"

	"github.com/mokiat/gocrane/internal/command/flag"
	"github.com/mokiat/gocrane/internal/project"
)

func Build() *cli.Command {
	var cfg buildConfig
	return &cli.Command{
		Name: "build",
		Flags: []cli.Flag{
			newVerboseFlag(&cfg.Verbose),
			newDirFlag(&cfg.Dirs),
			newDirExcludeFlag(&cfg.ExcludeDirs),
			newSourceFlag(&cfg.Sources),
			newSourceExcludeFlag(&cfg.ExcludeSources),
			newResourceFlag(&cfg.Resources),
			newResourceExcludeFlag(&cfg.ExcludeResources),
			newMainFlag(&cfg.MainDir),
			newBinaryFlag(&cfg.BinaryFile, true),
			newBuildArgs(&cfg.BuildArgs),
		},
		Action: func(c *cli.Context) error {
			return build(c.Context, cfg)
		},
	}
}

type buildConfig struct {
	Verbose          bool
	Dirs             cli.StringSlice
	ExcludeDirs      cli.StringSlice
	Sources          cli.StringSlice
	ExcludeSources   cli.StringSlice
	Resources        cli.StringSlice
	ExcludeResources cli.StringSlice
	MainDir          string
	BinaryFile       string
	BuildArgs        flag.ShlexStringSlice
}

func build(ctx context.Context, cfg buildConfig) error {
	log.Println("building binary...")
	builder := project.NewBuilder(cfg.MainDir, cfg.BuildArgs.Value())
	if err := builder.Build(ctx, cfg.BinaryFile); err != nil {
		return fmt.Errorf("failed to build binary: %w", err)
	}
	log.Println("binary successfully built.")

	log.Println("analyzing project...")
	layout := project.Explore(
		cfg.Dirs.Value(),
		cfg.ExcludeDirs.Value(),
		cfg.Sources.Value(),
		cfg.ExcludeSources.Value(),
		cfg.Resources.Value(),
		cfg.ExcludeResources.Value(),
	)
	log.Println("project successfully analyzed...")
	if cfg.Verbose {
		layout.PrintToLog()
	}

	log.Println("calculating digest...")
	dig, err := layout.Digest()
	if err != nil {
		return fmt.Errorf("failed to calculate digest: %w", err)
	}
	log.Printf("digest: %s", dig)

	log.Println("persisting digest...")
	digestFile := fmt.Sprintf("%s.dig", cfg.BinaryFile)
	if err := project.WriteDigest(digestFile, dig); err != nil {
		return fmt.Errorf("failed to write digest: %w", err)
	}
	log.Println("digest successfully persisted.")

	return nil
}
