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
	log.Println("Building binary...")
	builder := project.NewBuilder(cfg.MainDir, cfg.BuildArgs.Value())
	if err := builder.Build(ctx, cfg.BinaryFile); err != nil {
		return fmt.Errorf("failed to build binary: %w", err)
	}

	log.Println("Preparing filtering...")
	watchFilter, err := buildFilterTree(cfg.Dirs.Value(), cfg.ExcludeDirs.Value())
	if err != nil {
		return fmt.Errorf("problem with dir rules: %w", err)
	}
	sourceFilter, err := buildFilterTree(cfg.Sources.Value(), cfg.ExcludeSources.Value())
	if err != nil {
		return fmt.Errorf("problem with source rules: %w", err)
	}
	resourceFilter, err := buildFilterTree(cfg.Resources.Value(), cfg.ExcludeResources.Value())
	if err != nil {
		return fmt.Errorf("problem with resource rules: %w", err)
	}
	rootDirs := watchFilter.RootPaths()

	var summary *project.Summary
	if cfg.Verbose || cfg.BinaryFile != "" {
		log.Println("Analyzing project...")
		summary = project.Analyze(rootDirs, watchFilter, sourceFilter, resourceFilter)
	}
	if cfg.Verbose {
		printSummary(summary)
	}

	log.Println("Calculating current digest...")
	digest, err := calculateDigest(summary)
	if err != nil {
		return fmt.Errorf("failed to calculate digest: %w", err)
	}
	log.Printf("Digest: %s", digest)

	log.Println("Persisting digest...")
	digestFile := fmt.Sprintf("%s.dig", cfg.BinaryFile)
	if err := project.SaveDigestFile(digestFile, digest); err != nil {
		return fmt.Errorf("failed to write digest: %w", err)
	}
	log.Println("Digest successfully persisted.")

	log.Println("Done.")
	return nil
}
