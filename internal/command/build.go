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
			newIncludesFlag(&cfg.Includes),
			newExcludesFlag(&cfg.Excludes),
			newSourcesFlag(&cfg.Sources),
			newResourcesFlag(&cfg.Resources),
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
	Verbose    bool
	Includes   cli.StringSlice
	Excludes   cli.StringSlice
	Sources    cli.StringSlice
	Resources  cli.StringSlice
	MainDir    string
	BinaryFile string
	BuildArgs  flag.ShlexStringSlice
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
		cfg.Includes.Value(),
		cfg.Excludes.Value(),
		cfg.Sources.Value(),
		cfg.Resources.Value(),
	)
	log.Println("project successfully analyzed...")
	if cfg.Verbose {
		logLayout(layout)
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

func logLayout(layout *project.Layout) {
	log.Printf("omitted %d files or folders", len(layout.Omitted))
	for file, err := range layout.Omitted {
		log.Printf("omitted: %s (%s)", file, err)
	}

	log.Printf("found %d directories to watch", len(layout.WatchDirs))
	for _, dir := range layout.WatchDirs {
		log.Printf("watch dir: %s", dir)
	}

	log.Printf("found %d files to use for digest", len(layout.SourceFiles))
	for _, file := range layout.SourceFiles {
		log.Printf("source file: %s", file)
	}
}
